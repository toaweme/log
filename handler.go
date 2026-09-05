package log

import (
	"context"
	"log/slog"
)

// logger wraps *slog.Logger to implement Logger, adding the Trace helper.
type logger struct {
	slog *slog.Logger
}

var _ Logger = (*logger)(nil)

// With returns a logger that adds args to every subsequent record.
func (l *logger) With(args ...any) Logger {
	return &logger{slog: l.slog.With(args...)}
}

// WithGroup returns a logger that nests subsequent attributes under name.
func (l *logger) WithGroup(name string) Logger {
	return &logger{slog: l.slog.WithGroup(name)}
}

// Enabled reports whether the underlying handler emits records at level.
func (l *logger) Enabled(level slog.Level) bool {
	return l.slog.Enabled(context.Background(), level)
}

func (l *logger) Error(msg string, args ...any) { l.slog.Error(msg, args...) }
func (l *logger) Info(msg string, args ...any)  { l.slog.Info(msg, args...) }
func (l *logger) Debug(msg string, args ...any) { l.slog.Debug(msg, args...) }
func (l *logger) Warn(msg string, args ...any)  { l.slog.Warn(msg, args...) }

// Trace logs at the custom LevelTrace.
func (l *logger) Trace(msg string, args ...any) {
	l.slog.Log(context.Background(), LevelTrace, msg, args...)
}

// Slog returns the wrapped *slog.Logger.
func (l *logger) Slog() *slog.Logger { return l.slog }

// discardHandler drops every record. It reports Enabled false so callers skip
// building records at all.
type discardHandler struct{}

var _ slog.Handler = discardHandler{}

func (discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (h discardHandler) WithAttrs([]slog.Attr) slog.Handler      { return h }
func (h discardHandler) WithGroup(string) slog.Handler           { return h }
