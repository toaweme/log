package log

import (
	"log/slog"
	"os"
	"sync"
)

var (
	Level  = new(slog.LevelVar)
	Logger *slog.Logger
	mu     sync.RWMutex
)

func init() {
	Level.Set(slog.LevelDebug)
	SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: Level})))
}

func SetLogger(newLogger *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	Logger = newLogger
}

func SetLevel(lvl slog.Level) {
	mu.Lock()
	defer mu.Unlock()
	Level.Set(lvl)
}

func GetLevel() *slog.LevelVar {
	mu.RLock()
	defer mu.RUnlock()
	return Level
}

var (
	Error = func(msg string, args ...any) { Logger.Error(msg, args...) }
	Info  = func(msg string, args ...any) { Logger.Info(msg, args...) }
	Debug = func(msg string, args ...any) { Logger.Debug(msg, args...) }
	Warn  = func(msg string, args ...any) { Logger.Warn(msg, args...) }
)
