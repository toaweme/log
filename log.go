// Package log is a small, opinionated slog extension.
// Composable handlers (filtering, fan-out, per-logger level), two custom levels (TRACE, FATAL).
//
// Note: Fatal logs a FATAL record and returns. It does NOT call os.Exit; the
// decision to exit (and to flush or ship logs first) stays with the caller.
package log

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

// Logger is the extended slog contract: a slog.Handler plus level-aware helpers
// and the custom Trace/Fatal levels.
type Logger interface {
	slog.Handler
	With(args ...any) Logger
	// WithLevel returns a logger with a new minimum level, preserving the
	// underlying outputs, format, and attributes.
	WithLevel(level slog.Level) Logger
	Trace(msg string, args ...any)
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	// Fatal logs at LevelFatal and returns; it does not exit the process.
	Fatal(msg string, args ...any)
	// Slog returns the wrapped *slog.Logger as an escape hatch.
	Slog() *slog.Logger
}

const (
	// LevelTrace sits below slog.LevelDebug for the noisiest diagnostics.
	LevelTrace = slog.Level(-8)
	// LevelFatal sits above slog.LevelError. Logging at this level does not
	// exit the process; it only emits a FATAL record.
	LevelFatal = slog.Level(12)
)

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

// renameLevels renders the custom TRACE/FATAL levels with their names instead
// of slog's numeric fallback (e.g. "DEBUG-4").
func renameLevels(_ []string, a slog.Attr) slog.Attr {
	if a.Key != slog.LevelKey {
		return a
	}
	lvl, ok := a.Value.Any().(slog.Level)
	if !ok {
		return a
	}
	if name, exists := levelNames[lvl]; exists {
		a.Value = slog.StringValue(name)
	}
	return a
}

// Options returns slog.HandlerOptions wired to level with the custom-level name
// rendering, for building raw handlers passed to Output.
func Options(level slog.Leveler) *slog.HandlerOptions {
	return &slog.HandlerOptions{Level: level, ReplaceAttr: renameLevels}
}

// Option configures a Logger built by New.
type Option func(*builder)

type builder struct {
	level   *slog.LevelVar
	outputs []func(*slog.LevelVar) slog.Handler
	filters []Filter
}

// WithText adds a text handler writing to w.
func WithText(w io.Writer) Option {
	return func(b *builder) {
		b.outputs = append(b.outputs, func(lv *slog.LevelVar) slog.Handler {
			return slog.NewTextHandler(w, Options(lv))
		})
	}
}

// WithJSON adds a JSON handler writing to w. Pass a rotating writer (e.g. a
// lumberjack.Logger) here to keep that dependency out of this module.
func WithJSON(w io.Writer) Option {
	return func(b *builder) {
		b.outputs = append(b.outputs, func(lv *slog.LevelVar) slog.Handler {
			return slog.NewJSONHandler(w, Options(lv))
		})
	}
}

// WithOutput adds an arbitrary slog.Handler (a memory sink, an exporter, ...).
// The handler controls its own level; WithLevel does not affect it.
func WithOutput(h slog.Handler) Option {
	return func(b *builder) {
		b.outputs = append(b.outputs, func(*slog.LevelVar) slog.Handler { return h })
	}
}

// WithLevel sets the minimum level for the Text and JSON outputs (default Debug).
func WithLevel(level slog.Level) Option {
	return func(b *builder) { b.level.Set(level) }
}

// WithFilters wraps the assembled outputs in a FilteredLogger.
func WithFilters(filters ...Filter) Option {
	return func(b *builder) { b.filters = append(b.filters, filters...) }
}

// New assembles a Logger from the given outputs, level, and filters. With no
// outputs it writes text to stdout at Debug.
func New(opts ...Option) Logger {
	b := &builder{level: new(slog.LevelVar)}
	b.level.Set(slog.LevelDebug)
	for _, opt := range opts {
		opt(b)
	}

	if len(b.outputs) == 0 {
		b.outputs = append(b.outputs, func(lv *slog.LevelVar) slog.Handler {
			return slog.NewTextHandler(os.Stdout, Options(lv))
		})
	}

	handlers := make([]slog.Handler, len(b.outputs))
	for i, build := range b.outputs {
		handlers[i] = build(b.level)
	}

	var h slog.Handler
	if len(handlers) == 1 {
		h = handlers[0]
	} else {
		h = NewMultiHandler(handlers...)
	}
	if len(b.filters) > 0 {
		h = NewFilteredLogger(h, b.filters...)
	}

	return Wrap(slog.New(h))
}

// Wrap adopts an existing *slog.Logger as a Logger.
func Wrap(l *slog.Logger) Logger {
	return &logger{slog: l}
}

var (
	mu          sync.RWMutex
	globalLevel = new(slog.LevelVar)
	defaultLog  Logger
)

func init() {
	globalLevel.Set(slog.LevelDebug)
	defaultLog = Wrap(slog.New(slog.NewTextHandler(os.Stdout, Options(globalLevel))))
}

// Default returns the process-wide Logger backing the package-level helpers.
func Default() Logger {
	mu.RLock()
	defer mu.RUnlock()
	return defaultLog
}

// SetDefault replaces the process-wide Logger.
func SetDefault(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	defaultLog = l
}

// SetLevel sets the minimum level of the default logger built in init. It has
// no effect after SetDefault replaces the default with a logger of your own.
func SetLevel(level slog.Level) {
	globalLevel.Set(level)
}

// Level returns the default logger's minimum level.
func Level() slog.Level {
	return globalLevel.Level()
}

// Trace logs at LevelTrace on the default logger.
func Trace(msg string, args ...any) { Default().Trace(msg, args...) }

// Debug logs at slog.LevelDebug on the default logger.
func Debug(msg string, args ...any) { Default().Debug(msg, args...) }

// Info logs at slog.LevelInfo on the default logger.
func Info(msg string, args ...any) { Default().Info(msg, args...) }

// Warn logs at slog.LevelWarn on the default logger.
func Warn(msg string, args ...any) { Default().Warn(msg, args...) }

// Error logs at slog.LevelError on the default logger.
func Error(msg string, args ...any) { Default().Error(msg, args...) }

// Fatal logs at LevelFatal on the default logger; it does not exit the process.
func Fatal(msg string, args ...any) { Default().Fatal(msg, args...) }
