// Package log is a small, opinionated slog extension.
// It adds a TRACE level below DEBUG, a fan-out handler, a silent logger, and one
// process-wide level that every logger this package builds reads from.
package log

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Logger is the logging contract this package hands around.
// It stays deliberately narrow so a test double is a handful of methods.
type Logger interface {
	// With returns a logger that adds args to every subsequent record.
	With(args ...any) Logger
	// WithGroup returns a logger that nests subsequent attributes under name.
	WithGroup(name string) Logger
	// Enabled reports whether a record at level would be emitted.
	Enabled(level slog.Level) bool
	Trace(msg string, args ...any)
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	// Slog returns the wrapped *slog.Logger as an escape hatch.
	Slog() *slog.Logger
}

// LevelTrace sits below slog.LevelDebug for the noisiest diagnostics.
const LevelTrace = slog.Level(-8)

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
}

// ErrUnknownLevel reports a level name ParseLevel does not recognise.
var ErrUnknownLevel = errors.New("unknown log level")

// ParseLevel turns a level name into a slog.Level, case-insensitively.
// It accepts trace, debug, info, warn and error, and returns an error for
// anything else so the caller applies its own default.
func ParseLevel(name string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "trace":
		return LevelTrace, nil
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}
	return 0, fmt.Errorf("failed to parse log level %q: %w", name, ErrUnknownLevel)
}

// renameLevels renders the custom TRACE level with its name instead of slog's
// numeric fallback (e.g. "DEBUG-4").
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

// HandlerOptions returns slog.HandlerOptions wired to level with the custom-level
// name rendering, for building raw handlers passed to WithOutput.
func HandlerOptions(level slog.Leveler) *slog.HandlerOptions {
	return &slog.HandlerOptions{Level: level, ReplaceAttr: renameLevels}
}

// Option configures a Logger built by New.
type Option func(*builder)

type builder struct {
	outputs []slog.Handler
}

// WithText adds a text handler writing to w at the process-wide level.
func WithText(w io.Writer) Option {
	return func(b *builder) {
		b.outputs = append(b.outputs, slog.NewTextHandler(w, HandlerOptions(level)))
	}
}

// WithJSON adds a JSON handler writing to w at the process-wide level.
// Pass a rotating writer (e.g. a lumberjack.Logger) here to keep that dependency
// out of this module.
func WithJSON(w io.Writer) Option {
	return func(b *builder) {
		b.outputs = append(b.outputs, slog.NewJSONHandler(w, HandlerOptions(level)))
	}
}

// WithOutput adds an arbitrary slog.Handler (a memory sink, an exporter, ...).
// The handler carries whatever level it was built with, so Level describes it
// only when it was built with HandlerOptions(log.Level()).
func WithOutput(h slog.Handler) Option {
	return func(b *builder) { b.outputs = append(b.outputs, h) }
}

// WithLevel sets the process-wide level, the same one SetLevel moves and Level
// reports. There is one level for the whole process, so building a logger with
// this option changes what every other logger from this package emits.
func WithLevel(l slog.Level) Option {
	return func(*builder) { SetLevel(l) }
}

// New assembles a Logger from the given outputs.
// With no outputs it writes text to stdout.
func New(opts ...Option) Logger {
	b := &builder{}
	for _, opt := range opts {
		opt(b)
	}

	if len(b.outputs) == 0 {
		b.outputs = append(b.outputs, slog.NewTextHandler(os.Stdout, HandlerOptions(level)))
	}

	h := b.outputs[0]
	if len(b.outputs) > 1 {
		h = NewMultiHandler(b.outputs...)
	}

	return Wrap(slog.New(h))
}

// Wrap adopts an existing *slog.Logger as a Logger.
func Wrap(l *slog.Logger) Logger {
	return &logger{slog: l}
}

// Discard returns a Logger that drops every record and reports Enabled false at
// every level, so a caller can hold one instead of guarding a nil logger.
func Discard() Logger {
	return Wrap(slog.New(discardHandler{}))
}

var (
	mu    sync.Mutex
	level = newLevelVar(slog.LevelDebug)
	// defaultLog is built on first use of Default, so importing this package
	// opens no writer.
	defaultLog Logger
)

func newLevelVar(l slog.Level) *slog.LevelVar {
	lv := new(slog.LevelVar)
	lv.Set(l)
	return lv
}

// Default returns the process-wide Logger backing the package-level helpers,
// building a stdout text logger the first time it is asked for one.
func Default() Logger {
	mu.Lock()
	defer mu.Unlock()
	if defaultLog == nil {
		defaultLog = Wrap(slog.New(slog.NewTextHandler(os.Stdout, HandlerOptions(level))))
	}
	return defaultLog
}

// SetDefault replaces the process-wide Logger.
func SetDefault(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	defaultLog = l
}

// SetLevel moves the process-wide minimum level.
// Every logger New builds reads this value, so the change applies to all of them.
func SetLevel(l slog.Level) {
	level.Set(l)
}

// Level reports the process-wide minimum level.
func Level() slog.Level {
	return level.Level()
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
