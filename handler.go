package log

import (
	"context"
	"log/slog"
)

// logger wraps *slog.Logger to implement Logger, adding the Trace/Fatal helpers
// and a per-logger WithLevel.
type logger struct {
	slog *slog.Logger
}

var _ Logger = (*logger)(nil)

// With returns a logger that adds args to every subsequent record.
func (l *logger) With(args ...any) Logger {
	return &logger{slog: l.slog.With(args...)}
}

// WithLevel returns a logger whose minimum level is level, preserving the
// underlying handler's writer, format, and attributes. It wraps the existing
// handler rather than rebuilding it, so a custom output is not lost.
func (l *logger) WithLevel(level slog.Level) Logger {
	base := l.slog.Handler()
	// unwrap a previous level wrapper so repeated calls don't nest.
	if lh, ok := base.(*levelHandler); ok {
		base = lh.Handler
	}
	return &logger{slog: slog.New(&levelHandler{level: level, Handler: base})}
}

// Enabled reports whether the underlying handler emits records at level.
func (l *logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.slog.Enabled(ctx, level)
}

// Handle forwards the record to the underlying handler.
func (l *logger) Handle(ctx context.Context, record slog.Record) error {
	return l.slog.Handler().Handle(ctx, record)
}

// WithAttrs returns the underlying handler with attrs applied.
func (l *logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return l.slog.Handler().WithAttrs(attrs)
}

// WithGroup returns the underlying handler with the named group applied.
func (l *logger) WithGroup(name string) slog.Handler {
	return l.slog.Handler().WithGroup(name)
}

func (l *logger) Error(msg string, args ...any) { l.slog.Error(msg, args...) }
func (l *logger) Info(msg string, args ...any)  { l.slog.Info(msg, args...) }
func (l *logger) Debug(msg string, args ...any) { l.slog.Debug(msg, args...) }
func (l *logger) Warn(msg string, args ...any)  { l.slog.Warn(msg, args...) }

// Trace logs at the custom LevelTrace.
func (l *logger) Trace(msg string, args ...any) {
	l.slog.Log(context.Background(), LevelTrace, msg, args...)
}

// Fatal logs at the custom LevelFatal. It does not exit the process.
func (l *logger) Fatal(msg string, args ...any) {
	l.slog.Log(context.Background(), LevelFatal, msg, args...)
}

// Slog returns the wrapped *slog.Logger.
func (l *logger) Slog() *slog.Logger { return l.slog }

// levelHandler wraps a slog.Handler to enforce a minimum level while delegating
// everything else, so WithLevel can change the threshold without touching the
// output destination or format.
type levelHandler struct {
	level slog.Leveler
	slog.Handler
}

var _ slog.Handler = (*levelHandler)(nil)

// Enabled reports whether level meets the wrapper's minimum.
func (h *levelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// WithAttrs wraps the child handler's result, preserving the level threshold.
func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelHandler{level: h.level, Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup wraps the child handler's result, preserving the level threshold.
func (h *levelHandler) WithGroup(name string) slog.Handler {
	return &levelHandler{level: h.level, Handler: h.Handler.WithGroup(name)}
}
