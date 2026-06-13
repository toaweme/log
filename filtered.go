package log

import (
	"context"
	"log/slog"
	"strings"
	"sync"
)

// filterAction is what a matching Filter does to a record.
type filterAction int

const (
	// allow passes the record through unchanged (the zero value).
	allow filterAction = iota
	// deny drops the record so no downstream handler sees it.
	deny
	// shorten truncates the values of the configured attribute keys.
	shorten
)

// Filter selects log records and decides what happens to them, built fluently:
//
//	log.Deny().Below(slog.LevelInfo)
//	log.Deny().Attr("path", "/healthz*")
//	log.Shorten("body").Limit(200).Message("http response")
//
// A record must match every set criterion (level, message, attributes) for the
// filter's action to apply; criteria left unset are ignored.
type Filter struct {
	action      filterAction
	message     string
	attributes  map[string]string
	level       *slog.Level
	shortenKeys []string
	limit       int
}

// Deny starts a filter that drops matching records.
func Deny() Filter { return Filter{action: deny} }

// Allow starts a filter that passes matching records through unchanged.
func Allow() Filter { return Filter{action: allow} }

// Shorten starts a filter that truncates the given attribute keys on matching
// records. The default length limit is 100; change it with Limit.
func Shorten(keys ...string) Filter {
	return Filter{action: shorten, shortenKeys: keys, limit: 100}
}

// Message matches records whose message equals msg exactly.
func (f Filter) Message(msg string) Filter {
	f.message = msg
	return f
}

// Attr matches when the record's attribute key equals val. A val ending in "*"
// matches by prefix. The record message is available under the "msg" key.
func (f Filter) Attr(key, val string) Filter {
	attrs := make(map[string]string, len(f.attributes)+1)
	for k, v := range f.attributes {
		attrs[k] = v
	}
	attrs[key] = val
	f.attributes = attrs
	return f
}

// Below matches records strictly below level (e.g. Below(Info) matches Debug
// and Trace). Paired with Deny it acts as a level floor.
func (f Filter) Below(level slog.Level) Filter {
	f.level = &level
	return f
}

// Limit sets the Shorten length limit.
func (f Filter) Limit(n int) Filter {
	f.limit = n
	return f
}

// FilteredLogger is a slog.Handler that applies an ordered list of filters to
// each record before passing it to a wrapped handler. Filters run in order; the
// first Deny match drops the record, and Shorten matches rewrite attributes.
type FilteredLogger struct {
	handler slog.Handler
	filters []Filter
	mu      sync.RWMutex
}

var _ slog.Handler = (*FilteredLogger)(nil)

// NewFilteredLogger wraps handler with the given filters.
func NewFilteredLogger(handler slog.Handler, filters ...Filter) *FilteredLogger {
	return &FilteredLogger{
		handler: handler,
		filters: filters,
	}
}

// Enabled reports whether the wrapped handler emits records at level. Filters
// are evaluated in Handle, not here, since they can match on message or attrs.
func (f *FilteredLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return f.handler.Enabled(ctx, level)
}

// shortenMessage truncates msg to limit, using a trailing "..." when there is
// room for it (limit > 3).
func shortenMessage(msg string, limit int) string {
	if len(msg) <= limit {
		return msg
	}
	if limit <= 3 {
		return msg[:limit]
	}
	return msg[:limit-3] + "..."
}

// Handle applies each matching filter to the record and forwards the result to
// the wrapped handler. A matching Deny returns early and drops the record.
func (f *FilteredLogger) Handle(ctx context.Context, record slog.Record) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, filter := range f.filters {
		if !f.matchesFilter(record, filter) {
			continue
		}

		switch filter.action {
		case allow:
			// pass through unchanged.

		case deny:
			return nil

		case shorten:
			// build a set of keys to shorten once per filter
			if len(filter.shortenKeys) == 0 {
				break
			}
			keySet := make(map[string]struct{}, len(filter.shortenKeys))
			for _, k := range filter.shortenKeys {
				keySet[k] = struct{}{}
			}

			// reconstruct the record, replacing (not duplicating) matching attrs
			newRec := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
			record.Attrs(func(attr slog.Attr) bool {
				if _, ok := keySet[attr.Key]; ok {
					// replace value with shortened text
					newVal := shortenMessage(attr.Value.String(), filter.limit)
					newRec.AddAttrs(slog.String(attr.Key, newVal))
					return true
				}
				// keep original attr
				newRec.AddAttrs(attr)
				return true
			})

			// use the rebuilt record for downstream handler (no dupes)
			record = newRec
		}
	}

	return f.handler.Handle(ctx, record)
}

// WithAttrs returns a new FilteredLogger sharing the same filters, with attrs
// applied to the wrapped handler.
func (f *FilteredLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FilteredLogger{
		handler: f.handler.WithAttrs(attrs),
		filters: f.filters,
	}
}

// WithGroup returns a new FilteredLogger sharing the same filters, with the
// named group applied to the wrapped handler.
func (f *FilteredLogger) WithGroup(name string) slog.Handler {
	return &FilteredLogger{
		handler: f.handler.WithGroup(name),
		filters: f.filters,
	}
}

// AddFilter appends a filter; safe to call concurrently with logging.
func (f *FilteredLogger) AddFilter(filter Filter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.filters = append(f.filters, filter)
}

// SetFilters replaces the filter list; safe to call concurrently with logging.
func (f *FilteredLogger) SetFilters(filters []Filter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.filters = filters
}

// matchesFilter reports whether record satisfies every set criterion of filter.
func (f *FilteredLogger) matchesFilter(record slog.Record, filter Filter) bool {
	if filter.level != nil && *filter.level <= record.Level {
		return false
	}

	if filter.message != "" && record.Message != filter.message {
		return false
	}

	if len(filter.attributes) > 0 {
		recordAttrs := make(map[string]string)
		recordAttrs["msg"] = record.Message
		record.Attrs(func(attr slog.Attr) bool {
			recordAttrs[attr.Key] = attr.Value.String()
			return true
		})

		for filterKey, filterValue := range filter.attributes {
			if strings.HasSuffix(filterValue, "*") {
				valuePrefix := strings.TrimSuffix(filterValue, "*")
				matched := false
				for recordKey, recordValue := range recordAttrs {
					if filterKey == recordKey && strings.HasPrefix(recordValue, valuePrefix) {
						matched = true
						break
					}
				}
				if !matched {
					return false
				}
				continue
			}
			if recordAttrs[filterKey] != filterValue {
				return false
			}
		}
	}

	return true
}
