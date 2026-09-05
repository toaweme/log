package log

import (
	"bytes"
	"encoding/json"
	"errors"
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

// withLevel restores the process-wide level after a test moves it.
func withLevel(t *testing.T, l slog.Level) {
	t.Helper()
	prev := Level()
	t.Cleanup(func() { SetLevel(prev) })
	SetLevel(l)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			withGlobalLogger(t, slog.NewJSONHandler(&buf, HandlerOptions(LevelTrace)))

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

func Test_SetLevel_GatesLoggersBuiltByNew(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithText(&buf))
	withLevel(t, slog.LevelError)

	logger.Info("hidden")
	logger.Error("shown")

	out := buf.String()
	if strings.Contains(out, "hidden") {
		t.Fatalf("info was logged below the Error threshold: %q", out)
	}
	if !strings.Contains(out, "shown") {
		t.Fatalf("error record missing: %q", out)
	}
}

func Test_Level_AgreesWithNewWithLevelAndSetDefault(t *testing.T) {
	prev := Level()
	prevDefault := Default()
	t.Cleanup(func() {
		SetLevel(prev)
		SetDefault(prevDefault)
	})

	var buf bytes.Buffer
	SetDefault(New(WithText(&buf), WithLevel(slog.LevelWarn)))

	if got := Level(); got != slog.LevelWarn {
		t.Fatalf("Level() = %v, want %v", got, slog.LevelWarn)
	}

	Info("hidden")
	if buf.Len() != 0 {
		t.Fatalf("info leaked past the Warn threshold: %q", buf.String())
	}

	// SetLevel must still reach the logger New built.
	SetLevel(LevelTrace)
	if got := Level(); got != LevelTrace {
		t.Fatalf("Level() = %v, want %v", got, LevelTrace)
	}
	Trace("visible")
	if !strings.Contains(buf.String(), "visible") {
		t.Fatalf("SetLevel did not reach the installed logger: %q", buf.String())
	}
}

func Test_ParseLevel(t *testing.T) {
	tests := []struct {
		in      string
		want    slog.Level
		wantErr bool
	}{
		{in: "trace", want: LevelTrace},
		{in: "TRACE", want: LevelTrace},
		{in: "Debug", want: slog.LevelDebug},
		{in: "info", want: slog.LevelInfo},
		{in: " INFO ", want: slog.LevelInfo},
		{in: "warn", want: slog.LevelWarn},
		{in: "warning", want: slog.LevelWarn},
		{in: "eRRoR", want: slog.LevelError},
		{in: "", wantErr: true},
		{in: "fatal", wantErr: true},
		{in: "verbose", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseLevel(tt.in)
			if tt.wantErr {
				if !errors.Is(err, ErrUnknownLevel) {
					t.Fatalf("ParseLevel(%q) err = %v, want ErrUnknownLevel", tt.in, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseLevel(%q) returned %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("ParseLevel(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func Test_renameLevels_RendersCustomLevelNames(t *testing.T) {
	tests := []struct {
		level slog.Level
		want  string
	}{
		{LevelTrace, "TRACE"},
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

func Test_New_FansOutToEveryOutput(t *testing.T) {
	var text, jsonBuf bytes.Buffer
	withLevel(t, slog.LevelInfo)
	logger := New(WithText(&text), WithJSON(&jsonBuf))

	logger.Debug("below threshold")
	logger.Info("keep me", "k", "v")

	for _, buf := range []*bytes.Buffer{&text, &jsonBuf} {
		out := buf.String()
		if strings.Contains(out, "below threshold") {
			t.Fatalf("debug leaked past Info level: %q", out)
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
