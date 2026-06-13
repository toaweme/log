package log

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// withGlobalLogger swaps the default logger for the duration of a test and
// restores it afterwards, since the helpers operate on global state.
func withGlobalLogger(t *testing.T, h slog.Handler) {
	t.Helper()
	prev := Default()
	SetDefault(Wrap(slog.New(h)))
	t.Cleanup(func() { SetDefault(prev) })
}

func Test_PackageHelpers_RouteToDefaultLogger(t *testing.T) {
	tests := []struct {
		name      string
		log       func()
		wantMsg   string
		wantLevel string
	}{
		{"info", func() { Info("i") }, "i", "INFO"},
		{"error", func() { Error("e") }, "e", "ERROR"},
		{"debug", func() { Debug("d") }, "d", "DEBUG"},
		{"warn", func() { Warn("w") }, "w", "WARN"},
		{"trace", func() { Trace("t") }, "t", "TRACE"},
		{"fatal", func() { Fatal("f") }, "f", "FATAL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			withGlobalLogger(t, slog.NewJSONHandler(&buf, Options(LevelTrace)))

			tt.log()

			var rec map[string]any
			if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &rec); err != nil {
				t.Fatalf("unmarshal %q: %v", buf.String(), err)
			}
			if rec["msg"] != tt.wantMsg {
				t.Fatalf("msg = %v, want %q", rec["msg"], tt.wantMsg)
			}
			if rec["level"] != tt.wantLevel {
				t.Fatalf("level = %v, want %q", rec["level"], tt.wantLevel)
			}
		})
	}
}

func Test_SetLevel_GatesPackageHelpers(t *testing.T) {
	var buf bytes.Buffer
	// the default logger built in init is wired to globalLevel, so a handler
	// wired to the same var reproduces that behavior under SetLevel.
	withGlobalLogger(t, slog.NewTextHandler(&buf, Options(globalLevel)))

	prev := Level()
	t.Cleanup(func() { SetLevel(prev) })

	SetLevel(slog.LevelError)
	Info("hidden")
	Error("shown")

	out := buf.String()
	if strings.Contains(out, "hidden") {
		t.Fatalf("info was logged below the Error threshold: %q", out)
	}
	if !strings.Contains(out, "shown") {
		t.Fatalf("error record missing: %q", out)
	}
}

func Test_renameLevels_RendersCustomLevelNames(t *testing.T) {
	tests := []struct {
		level slog.Level
		want  string
	}{
		{LevelTrace, "TRACE"},
		{LevelFatal, "FATAL"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelError, "ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			in := slog.Any(slog.LevelKey, tt.level)
			got := renameLevels(nil, in)
			if got.Value.String() != tt.want {
				t.Fatalf("renameLevels level = %q, want %q", got.Value.String(), tt.want)
			}
		})
	}
}

func Test_renameLevels_LeavesNonLevelAttrsUntouched(t *testing.T) {
	in := slog.String("user", "bob")
	got := renameLevels(nil, in)
	if got.Key != "user" || got.Value.String() != "bob" {
		t.Fatalf("renameLevels mutated a non-level attr: %v", got)
	}
}

func Test_New_AssemblesOutputsAndFilters(t *testing.T) {
	var text, jsonBuf bytes.Buffer
	logger := New(
		WithText(&text),
		WithJSON(&jsonBuf),
		WithLevel(slog.LevelInfo),
		WithFilters(Deny().Message("drop me")),
	)

	logger.Debug("below threshold")
	logger.Info("drop me")
	logger.Info("keep me", "k", "v")

	for _, buf := range []*bytes.Buffer{&text, &jsonBuf} {
		out := buf.String()
		if strings.Contains(out, "below threshold") {
			t.Fatalf("debug leaked past Info level: %q", out)
		}
		if strings.Contains(out, "drop me") {
			t.Fatalf("denied record was emitted: %q", out)
		}
		if !strings.Contains(out, "keep me") {
			t.Fatalf("kept record missing from output: %q", out)
		}
	}

	// the JSON output must be valid JSON, proving fan-out built distinct handlers.
	var rec map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(jsonBuf.String())), &rec); err != nil {
		t.Fatalf("json output invalid: %q (%v)", jsonBuf.String(), err)
	}
	if rec["k"] != "v" {
		t.Fatalf("attr k = %v, want %q", rec["k"], "v")
	}
}
