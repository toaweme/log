package log

import (
	"context"
	"log/slog"
	"sync"
)

type Filter struct {
	Level   *slog.Level
	Message string
	// Attributes is a map of attribute key-value pairs to match against the log record
	// query map
	Attributes map[string]string
	// Options can be used to pass additional options for the filter
	// e.g., to specify a message length limit for shortening
	// shorten: 20
	Options map[string]map[string]any
	Action  FilterAction
}

var OptShorten = "shorten"
var OptShortenLimit = "shorten:limit"
var OptShortenKeys = "shorten:keys"

func (f Filter) LengthLimit() int {
	option := getOption(f.Options, string(OptShorten))
	if option == nil {
		return 0
	}
	return getIntOption(option, OptShortenLimit, 100)
}

func (f Filter) ShortenKeys() []string {
	option := getOption(f.Options, string(OptShorten))
	if option == nil {
		return nil
	}
	keys, ok := option[OptShortenKeys]
	if !ok {
		return nil
	}
	switch v := keys.(type) {
	case []string:
		return v
	}
	return nil
}

func getOption(options map[string]map[string]any, key string) map[string]any {
	if options == nil || len(options) == 0 {
		return nil
	}
	if val, ok := options[key]; ok {
		return val
	}
	return nil
}

// handle all number types
func getIntOption(options map[string]any, key string, def int) int {
	if options == nil || len(options) == 0 {
		return def
	}
	if val, ok := options[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int8:
			return int(v)
		case int16:
			return int(v)
		case int32:
			return int(v)
		case int64:
			return int(v)
		case uint:
			return int(v)
		case uint8:
			return int(v)
		case uint16:
			return int(v)
		case uint32:
			return int(v)
		case uint64:
			return int(v)
		case float32:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return def
}

type FilterAction int

const (
	Allow FilterAction = iota
	Deny
	Shorten
)

type FilteredLogger struct {
	handler slog.Handler
	filters []Filter
	mu      sync.RWMutex
}

func NewFilteredLogger(handler slog.Handler, filters ...Filter) *FilteredLogger {
	return &FilteredLogger{
		handler: handler,
		filters: filters,
	}
}

func (f *FilteredLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return f.handler.Enabled(ctx, level)
}

func shortenMessage(msg string, limit int) string {
	if len(msg) <= limit {
		return msg
	}
	if limit <= 3 {
		return msg[:limit]
	}
	return msg[:limit-3] + "..."
}
func (f *FilteredLogger) Handle(ctx context.Context, record slog.Record) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, filter := range f.filters {
		if !f.matchesFilter(record, filter) {
			continue
		}

		switch filter.Action {
		case Deny:
			return nil

		case Shorten:
			// build a set of keys to shorten once per filter
			keys := filter.ShortenKeys()
			if len(keys) == 0 {
				break
			}
			keySet := make(map[string]struct{}, len(keys))
			for _, k := range keys {
				keySet[k] = struct{}{}
			}

			// reconstruct the record, replacing (not duplicating) matching attrs
			newRec := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
			record.Attrs(func(attr slog.Attr) bool {
				if _, ok := keySet[attr.Key]; ok {
					// replace value with shortened text
					newVal := shortenMessage(attr.Value.String(), filter.LengthLimit())
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

func (f *FilteredLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FilteredLogger{
		handler: f.handler.WithAttrs(attrs),
		filters: f.filters,      // Share the same filters
		mu:      sync.RWMutex{}, // New mutex for the new handler
	}
}

func (f *FilteredLogger) WithGroup(name string) slog.Handler {
	return &FilteredLogger{
		handler: f.handler.WithGroup(name),
		filters: f.filters,      // Share the same filters
		mu:      sync.RWMutex{}, // New mutex for the new handler
	}
}

func (f *FilteredLogger) AddFilter(filter Filter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.filters = append(f.filters, filter)
}

func (f *FilteredLogger) SetFilters(filters []Filter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.filters = filters
}

func (f *FilteredLogger) matchesFilter(record slog.Record, filter Filter) bool {
	if filter.Level != nil && record.Level != *filter.Level {
		return false
	}

	if filter.Message != "" && record.Message != filter.Message {
		return false
	}

	if len(filter.Attributes) > 0 {
		recordAttrs := make(map[string]string)
		recordAttrs["msg"] = record.Message
		record.Attrs(func(attr slog.Attr) bool {
			recordAttrs[attr.Key] = attr.Value.String()
			return true
		})

		for key, value := range filter.Attributes {
			if recordAttrs[key] != value {
				return false
			}
		}
	}

	return true
}
