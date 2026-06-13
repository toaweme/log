package log

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

func Test_MultiHandler_Enabled(t *testing.T) {
	tests := []struct {
		name string
		mins []slog.Level
		at   slog.Level
		want bool
	}{
		{
			name: "enabled when any child is enabled",
			mins: []slog.Level{slog.LevelError, slog.LevelDebug},
			at:   slog.LevelInfo,
			want: true,
		},
		{
			name: "disabled only when every child is disabled",
			mins: []slog.Level{slog.LevelError, slog.LevelWarn},
			at:   slog.LevelInfo,
			want: false,
		},
		{
			name: "no handlers is disabled",
			mins: nil,
			at:   slog.LevelInfo,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handlers []slog.Handler
			for _, m := range tt.mins {
				handlers = append(handlers, newFakeHandler(m, nil))
			}
			mh := NewMultiHandler(handlers...)
			if got := mh.Enabled(context.Background(), tt.at); got != tt.want {
				t.Fatalf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_MultiHandler_Handle_DispatchesToEnabledOnly(t *testing.T) {
	low := newFakeHandler(slog.LevelDebug, nil)
	high := newFakeHandler(slog.LevelError, nil)
	mh := NewMultiHandler(low, high)

	if err := mh.Handle(context.Background(), newRecord(slog.LevelInfo, "hi")); err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	if got := low.messages(); len(got) != 1 || got[0] != "hi" {
		t.Fatalf("low handler messages = %v, want [hi]", got)
	}
	if got := high.messages(); len(got) != 0 {
		t.Fatalf("high handler messages = %v, want none (level gated out)", got)
	}
}

func Test_MultiHandler_Handle_JoinsErrors(t *testing.T) {
	errA := errors.New("handler a failed")
	errB := errors.New("handler b failed")

	a := newFakeHandler(slog.LevelDebug, errA)
	b := newFakeHandler(slog.LevelDebug, errB)
	ok := newFakeHandler(slog.LevelDebug, nil)
	mh := NewMultiHandler(a, ok, b)

	err := mh.Handle(context.Background(), newRecord(slog.LevelInfo, "hi"))
	if err == nil {
		t.Fatal("Handle() returned nil, want joined error")
	}
	if !errors.Is(err, errA) || !errors.Is(err, errB) {
		t.Fatalf("Handle() error = %v, want both errA and errB joined", err)
	}

	// a failing handler must not stop the others from receiving the record.
	if got := ok.messages(); len(got) != 1 {
		t.Fatalf("ok handler messages = %v, want it still received the record", got)
	}
}

func Test_MultiHandler_WithAttrsAndGroup(t *testing.T) {
	a := newRecHandler(LevelTrace)
	b := newRecHandler(LevelTrace)
	mh := NewMultiHandler(a, b)

	withAttrs := mh.WithAttrs([]slog.Attr{slog.String("svc", "api")})
	if _, ok := withAttrs.(*MultiHandler); !ok {
		t.Fatalf("WithAttrs() = %T, want *MultiHandler", withAttrs)
	}

	withGroup := mh.WithGroup("g")
	if _, ok := withGroup.(*MultiHandler); !ok {
		t.Fatalf("WithGroup() = %T, want *MultiHandler", withGroup)
	}
}
