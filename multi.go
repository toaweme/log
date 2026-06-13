package log

import (
	"context"
	"errors"
	"log/slog"
)

// MultiHandler fans a record out to several handlers, so one logger can write
// to multiple outputs (e.g. a text console and a JSON file) at once.
type MultiHandler struct {
	handlers []slog.Handler
}

var _ slog.Handler = (*MultiHandler)(nil)

// NewMultiHandler returns a handler that dispatches to each of handlers in order.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// WithAttrs returns a new MultiHandler with attrs applied to every child handler.
func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

// WithGroup returns a new MultiHandler with the group applied to every child.
func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

// Enabled reports whether any child handler is enabled for the level, so a
// record is dropped only when every output would discard it.
func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

// Handle dispatches the record to every enabled child handler and joins any
// errors, so one failing output does not stop the others.
func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
