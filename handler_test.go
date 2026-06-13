package log

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func Test_Logger_WithLevel_PreservesWriterAndFormat(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, Options(LevelTrace)))
	logger := Wrap(base)

	// raise the threshold to Error; the writer and JSON format must be kept.
	leveled := logger.WithLevel(slog.LevelError)

	leveled.Info("suppressed")
	leveled.Error("kept")

	out := buf.String()
	if strings.Contains(out, "suppressed") {
		t.Fatalf("info record was emitted despite Error threshold: %q", out)
	}
	if !strings.Contains(out, "kept") {
		t.Fatalf("error record missing; writer was not preserved: %q", out)
	}

	// the format must still be JSON (proving the handler was not rebuilt as text).
	var rec map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rec); err != nil {
		t.Fatalf("output is not JSON, format was lost: %q (%v)", out, err)
	}
	if rec["msg"] != "kept" {
		t.Fatalf("msg = %v, want %q", rec["msg"], "kept")
	}
}

func Test_Logger_WithLevel_CanWidenAgain(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewTextHandler(&buf, Options(LevelTrace)))
	logger := Wrap(base)

	narrowed := logger.WithLevel(slog.LevelError)
	widened := narrowed.WithLevel(LevelTrace)

	widened.Info("now visible")
	if !strings.Contains(buf.String(), "now visible") {
		t.Fatalf("re-widening did not re-enable Info: %q", buf.String())
	}
}

func Test_Logger_With_AddsAttributes(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, Options(LevelTrace)))
	logger := Wrap(base).With("svc", "api")

	logger.Info("hello")

	var rec map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rec["svc"] != "api" {
		t.Fatalf("svc = %v, want %q", rec["svc"], "api")
	}
}

func Test_Logger_CustomLevels(t *testing.T) {
	tests := []struct {
		name      string
		log       func(l Logger)
		wantLevel string
	}{
		{"trace", func(l Logger) { l.Trace("t") }, "TRACE"},
		{"fatal", func(l Logger) { l.Fatal("f") }, "FATAL"},
		{"info", func(l Logger) { l.Info("i") }, "INFO"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			base := slog.New(slog.NewJSONHandler(&buf, Options(LevelTrace)))
			logger := Wrap(base)

			tt.log(logger)

			var rec map[string]any
			if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &rec); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if rec["level"] != tt.wantLevel {
				t.Fatalf("level = %v, want %q", rec["level"], tt.wantLevel)
			}
		})
	}
}

func Test_levelHandler_Enabled(t *testing.T) {
	down := newRecHandler(LevelTrace)
	lh := &levelHandler{level: slog.LevelWarn, Handler: down}

	tests := []struct {
		level slog.Level
		want  bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, false},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}
	for _, tt := range tests {
		if got := lh.Enabled(context.Background(), tt.level); got != tt.want {
			t.Fatalf("Enabled(%v) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

func Test_Logger_Slog_RoundTrips(t *testing.T) {
	base := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
	logger := Wrap(base)
	if logger.Slog() != base {
		t.Fatal("Slog() did not return the wrapped *slog.Logger")
	}
}
