package log

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func Test_Logger_With_AddsAttributes(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, HandlerOptions(LevelTrace)))
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

func Test_Logger_WithGroup_NestsAttributes(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, HandlerOptions(LevelTrace)))
	logger := Wrap(base).WithGroup("req").With("id", "42")

	logger.Info("hello")

	var rec map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	group, ok := rec["req"].(map[string]any)
	if !ok {
		t.Fatalf("req group missing: %v", rec)
	}
	if group["id"] != "42" {
		t.Fatalf("req.id = %v, want %q", group["id"], "42")
	}
}

func Test_Logger_Enabled_FollowsTheHandler(t *testing.T) {
	logger := Wrap(slog.New(newRecHandler(slog.LevelWarn)))

	tests := []struct {
		level slog.Level
		want  bool
	}{
		{LevelTrace, false},
		{slog.LevelDebug, false},
		{slog.LevelInfo, false},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}
	for _, tt := range tests {
		if got := logger.Enabled(tt.level); got != tt.want {
			t.Fatalf("Enabled(%v) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

func Test_Logger_CustomLevels(t *testing.T) {
	tests := []struct {
		name      string
		log       func(l Logger)
		wantLevel string
	}{
		{"trace", func(l Logger) { l.Trace("t") }, "TRACE"},
		{"info", func(l Logger) { l.Info("i") }, "INFO"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			base := slog.New(slog.NewJSONHandler(&buf, HandlerOptions(LevelTrace)))
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

func Test_Logger_Slog_RoundTrips(t *testing.T) {
	base := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
	logger := Wrap(base)
	if logger.Slog() != base {
		t.Fatal("Slog() did not return the wrapped *slog.Logger")
	}
}

func Test_Discard_DropsEverythingAndDisablesAllLevels(t *testing.T) {
	l := Discard()

	levels := []slog.Level{LevelTrace, slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	for _, lv := range levels {
		if l.Enabled(lv) {
			t.Fatalf("Enabled(%v) = true, want false", lv)
		}
		if l.With("k", "v").Enabled(lv) {
			t.Fatalf("With().Enabled(%v) = true, want false", lv)
		}
		if l.WithGroup("g").Enabled(lv) {
			t.Fatalf("WithGroup().Enabled(%v) = true, want false", lv)
		}
	}

	// these must not panic and must produce no output.
	l.Info("i")
	l.Error("e")
	l.With("k", "v").Warn("w")
	l.Trace("t")
}
