package log

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

var (
	level  = new(slog.LevelVar)
	Logger *slog.Logger
	mu     sync.RWMutex
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

func init() {
	level.Set(slog.LevelDebug)
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := levelNames[level]
				if !exists {
					levelLabel = level.String()
				}
				
				a.Value = slog.StringValue(levelLabel)
			}
			
			return a
		},
	}
	SetLogger(slog.New(slog.NewTextHandler(os.Stdout, opts)))
}

func SetLogger(newLogger *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	Logger = newLogger
}

func SetLevel(lvl slog.Level) {
	mu.Lock()
	defer mu.Unlock()
	level.Set(lvl)
}

func GetLevel() *slog.LevelVar {
	mu.RLock()
	defer mu.RUnlock()
	return level
}

var (
	Error = func(msg string, args ...any) { Logger.Error(msg, args...) }
	Info  = func(msg string, args ...any) { Logger.Info(msg, args...) }
	Debug = func(msg string, args ...any) { Logger.Debug(msg, args...) }
	Warn  = func(msg string, args ...any) { Logger.Warn(msg, args...) }
	Trace = func(msg string, args ...any) { Logger.Log(context.Background(), LevelTrace, msg, args...) }
	Fatal = func(msg string, args ...any) { Logger.Log(context.Background(), LevelFatal, msg, args...) }
)
