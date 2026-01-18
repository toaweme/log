package log

import (
	"context"
	"log/slog"
	"os"
)

// ExtendedLogger wraps *slog.Logger to implement our Slog interface
type ExtendedLogger struct {
	logger *slog.Logger
	attrs  []any
}

func NewExtendedLogger(logger *slog.Logger) *ExtendedLogger {
	return &ExtendedLogger{
		logger: logger,
		attrs:  []any{},
	}
}

var _ Slog = (*ExtendedLogger)(nil)

func (l *ExtendedLogger) With(args ...any) Slog {
	newAttrs := make([]any, len(l.attrs)+len(args))
	copy(newAttrs, l.attrs)
	copy(newAttrs[len(l.attrs):], args)

	return &ExtendedLogger{
		logger: l.logger.With(args...),
		attrs:  newAttrs,
	}
}

// WithLevel creates a new logger with a different level
func (l *ExtendedLogger) WithLevel(level slog.Level) Slog {
	// Create new handler with the specified level
	var handler slog.Handler

	// Try to determine the handler type and recreate it
	// This is a simple approach - you might want to make this more sophisticated
	switch l.logger.Handler().(type) {
	case *slog.JSONHandler:
		handler = slog.NewJSONHandler(os.Stdout, CreateLoggerOptions(level))
	default:
		handler = slog.NewTextHandler(os.Stdout, CreateLoggerOptions(level))
	}

	// Create new logger with the new handler
	newLogger := slog.New(handler)

	// Re-apply any stored attributes
	if len(l.attrs) > 0 {
		newLogger = newLogger.With(l.attrs...)
	}

	return &ExtendedLogger{
		logger: newLogger,
		attrs:  l.attrs,
	}
}

func (l *ExtendedLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.logger.Enabled(ctx, level)
}

func (l *ExtendedLogger) Handle(ctx context.Context, record slog.Record) error {
	return l.logger.Handler().Handle(ctx, record)
}

func (l *ExtendedLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return l.logger.Handler().WithAttrs(attrs)
}

func (l *ExtendedLogger) WithGroup(name string) slog.Handler {
	return l.logger.Handler().WithGroup(name)
}

func (l *ExtendedLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *ExtendedLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *ExtendedLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *ExtendedLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *ExtendedLogger) Trace(msg string, args ...any) {
	l.logger.Log(context.Background(), LevelTrace, msg, args...)
}

func (l *ExtendedLogger) Fatal(msg string, args ...any) {
	l.logger.Log(context.Background(), LevelFatal, msg, args...)
}

func (l *ExtendedLogger) Logger() *slog.Logger {
	return l.logger
}
