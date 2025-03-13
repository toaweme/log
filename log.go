package log

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
	
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	level  = new(slog.LevelVar)
	Logger *slog.Logger
	mu     sync.RWMutex
)

// Internal function to set the global logger
func setLogger(newLogger *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	Logger = newLogger
}

func init() {
	level.Set(slog.LevelDebug)
	setLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}

var (
	Error = func(msg string, args ...any) { Logger.Error(msg, args...) }
	Info  = func(msg string, args ...any) { Logger.Info(msg, args...) }
	Debug = func(msg string, args ...any) { Logger.Debug(msg, args...) }
	Warn  = func(msg string, args ...any) { Logger.Warn(msg, args...) }
)

func SetLevel(lvl slog.Level) {
	level.Set(lvl)
}

func SetupGlobalLogger(appLogDir string) func() {
	// Default `stdout` pretty logging
	stdoutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	
	// Log filename format
	format := time.Now().Format("20060102T150405")
	logFileName := filepath.Join(appLogDir, "awee-"+runtime.GOOS+"-"+format+".log")
	
	// File logger setup (JSON format)
	fileHandler := slog.NewJSONHandler(&lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    2,  // MB
		MaxBackups: 5,  // Number of backups
		MaxAge:     28, // Days
		Compress:   false,
	}, &slog.HandlerOptions{Level: level})
	
	setLogger(slog.New(&MultiHandler{handlers: []slog.Handler{stdoutHandler, fileHandler}}))
	
	return func() {
		Info("Shutting down logging")
	}
}

// MultiHandler enables logging to multiple outputs
type MultiHandler struct {
	handlers []slog.Handler
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		_ = h.Handle(ctx, r) // Ignore errors
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}
