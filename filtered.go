package log

import (
	"context"
	"log/slog"
	"sync"
)

type Filter struct {
	Messages []FilterMessage
}

type FilterMessage struct {
	Level   *slog.Level
	Message string
}

type FilteredLogger struct {
	logger *slog.Logger
	filter Filter

	mu *sync.RWMutex
}

func NewFilteredLogger(logger *slog.Logger, filter Filter) *FilteredLogger {
	return &FilteredLogger{
		logger: logger,
		filter: filter,

		mu: &sync.RWMutex{},
	}
}

func (f *FilteredLogger) Filter(filter Filter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.filter = filter
}

func (f *FilteredLogger) Error(msg string, args ...any) {
	f.log(slog.LevelError, msg, args...)
}

func (f *FilteredLogger) Info(msg string, args ...any) {
	f.log(slog.LevelInfo, msg, args...)
}

func (f *FilteredLogger) Debug(msg string, args ...any) {
	f.log(slog.LevelDebug, msg, args...)
}

func (f *FilteredLogger) Warn(msg string, args ...any) {
	f.log(slog.LevelWarn, msg, args...)
}

func (f *FilteredLogger) Trace(msg string, args ...any) {
	f.log(LevelTrace, msg, args...)
}

func (f *FilteredLogger) Fatal(msg string, args ...any) {
	f.log(LevelFatal, msg, args...)
}

func (f *FilteredLogger) log(level slog.Level, msg string, args ...any) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, filterMsg := range f.filter.Messages {
		if filterMsg.Level != nil && *filterMsg.Level >= level {
			f.logger.Log(context.Background(), level, msg, args...)
		}
		if filterMsg.Message != "" && filterMsg.Message == msg {
			f.logger.Log(context.Background(), level, msg, args...)
			return
		}
	}
}
