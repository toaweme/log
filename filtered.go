package log

import (
	"context"
	"log/slog"
	"sync"
)

type Filter struct {
	Level      *slog.Level
	Message    string
	Attributes map[string]string
	Action     FilterAction
}

type FilterAction int

const (
	Allow FilterAction = iota
	Deny
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

func (f *FilteredLogger) Handle(ctx context.Context, record slog.Record) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, filter := range f.filters {
		if f.matchesFilter(record, filter) {
			if filter.Action == Deny {
				return nil
			}
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
