package log

import (
	"context"
	"log/slog"
)

// ExtendedLogger wraps *slog.Logger to implement our Slog interface
type ExtendedLogger struct {
	logger *slog.Logger
}

func NewExtendedLogger(logger *slog.Logger) *ExtendedLogger {
	return &ExtendedLogger{logger: logger}
}

var _ Slog = (*ExtendedLogger)(nil)

func (l *ExtendedLogger) With(args ...any) Slog {
	return &ExtendedLogger{
		logger: l.logger.With(args...),
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
