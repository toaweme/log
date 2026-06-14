package log

import (
	"context"
	"log/slog"
	"sync"
	"testing"
)

func Test_FilteredLogger_matchesFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		record slog.Record
		want   bool
	}{
		{
			name:   "empty filter matches anything",
			filter: Allow(),
			record: newRecord(slog.LevelInfo, "hello"),
			want:   true,
		},
		{
			name:   "level matches records strictly below it",
			filter: Allow().Below(slog.LevelInfo),
			record: newRecord(slog.LevelDebug, "hello"),
			want:   true,
		},
		{
			name:   "level does not match record at the same level",
			filter: Allow().Below(slog.LevelInfo),
			record: newRecord(slog.LevelInfo, "hello"),
			want:   false,
		},
		{
			name:   "level does not match record above it",
			filter: Allow().Below(slog.LevelInfo),
			record: newRecord(slog.LevelError, "hello"),
			want:   false,
		},
		{
			name:   "exact message match",
			filter: Allow().Message("ping"),
			record: newRecord(slog.LevelInfo, "ping"),
			want:   true,
		},
		{
			name:   "message mismatch",
			filter: Allow().Message("ping"),
			record: newRecord(slog.LevelInfo, "pong"),
			want:   false,
		},
		{
			name:   "attribute exact match",
			filter: Allow().Attr("user", "bob"),
			record: newRecord(slog.LevelInfo, "hi", "user", "bob"),
			want:   true,
		},
		{
			name:   "attribute value mismatch",
			filter: Allow().Attr("user", "bob"),
			record: newRecord(slog.LevelInfo, "hi", "user", "alice"),
			want:   false,
		},
		{
			name:   "missing attribute key does not match",
			filter: Allow().Attr("user", "bob"),
			record: newRecord(slog.LevelInfo, "hi"),
			want:   false,
		},
		{
			name:   "wildcard prefix match",
			filter: Allow().Attr("path", "/api/*"),
			record: newRecord(slog.LevelInfo, "hi", "path", "/api/users"),
			want:   true,
		},
		{
			name:   "wildcard prefix mismatch",
			filter: Allow().Attr("path", "/api/*"),
			record: newRecord(slog.LevelInfo, "hi", "path", "/web/home"),
			want:   false,
		},
		{
			name:   "match against synthetic msg attribute",
			filter: Allow().Attr("msg", "health*"),
			record: newRecord(slog.LevelInfo, "healthcheck ok"),
			want:   true,
		},
		{
			name:   "bare wildcard matches any present value",
			filter: Allow().Attr("path", "*"),
			record: newRecord(slog.LevelInfo, "hi", "path", "/anything"),
			want:   true,
		},
		{
			name:   "bare wildcard does not match an absent key",
			filter: Allow().Attr("path", "*"),
			record: newRecord(slog.LevelInfo, "hi", "other", "x"),
			want:   false,
		},
		{
			name:   "wildcard prefix does not match an absent key",
			filter: Allow().Attr("path", "/api/*"),
			record: newRecord(slog.LevelInfo, "hi"),
			want:   false,
		},
		{
			name:   "wildcard matches a present empty value when prefix is empty",
			filter: Allow().Attr("path", "*"),
			record: newRecord(slog.LevelInfo, "hi", "path", ""),
			want:   true,
		},
		{
			name:   "all criteria must match (one fails)",
			filter: Allow().Message("ping").Attr("user", "bob"),
			record: newRecord(slog.LevelInfo, "ping", "user", "alice"),
			want:   false,
		},
		{
			name:   "all criteria must match (all pass)",
			filter: Allow().Below(slog.LevelWarn).Message("ping").Attr("user", "bob"),
			record: newRecord(slog.LevelInfo, "ping", "user", "bob"),
			want:   true,
		},
	}

	f := &FilteredLogger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.matchesFilter(tt.record, tt.filter)
			if got != tt.want {
				t.Fatalf("matchesFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Filter_Attr_DoesNotMutateSharedBase(t *testing.T) {
	base := Allow().Attr("a", "1")
	withB := base.Attr("b", "2")

	if len(base.attributes) != 1 {
		t.Fatalf("base filter gained attrs from a derived filter: %v", base.attributes)
	}
	if len(withB.attributes) != 2 {
		t.Fatalf("derived filter missing attrs: %v", withB.attributes)
	}
}

func Test_FilteredLogger_Handle_Deny(t *testing.T) {
	down := newRecHandler(LevelTrace)
	fl := NewFilteredLogger(down, Deny().Message("secret"))

	if err := fl.Handle(context.Background(), newRecord(slog.LevelInfo, "secret")); err != nil {
		t.Fatalf("Handle() denied record returned error: %v", err)
	}
	if err := fl.Handle(context.Background(), newRecord(slog.LevelInfo, "public")); err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	seen := down.seen()
	if len(seen) != 1 {
		t.Fatalf("downstream saw %d records, want 1", len(seen))
	}
	if seen[0].Message != "public" {
		t.Fatalf("downstream message = %q, want %q", seen[0].Message, "public")
	}
}

func Test_FilteredLogger_Handle_DenyBelowLevel(t *testing.T) {
	down := newRecHandler(LevelTrace)
	fl := NewFilteredLogger(down, Deny().Below(slog.LevelInfo))

	_ = fl.Handle(context.Background(), newRecord(slog.LevelDebug, "noise"))
	_ = fl.Handle(context.Background(), newRecord(slog.LevelInfo, "kept"))

	seen := down.seen()
	if len(seen) != 1 || seen[0].Message != "kept" {
		t.Fatalf("Deny().Below(Info) did not act as a floor: %v", seen)
	}
}

func Test_FilteredLogger_Handle_Allow(t *testing.T) {
	down := newRecHandler(LevelTrace)
	fl := NewFilteredLogger(down, Allow())

	if err := fl.Handle(context.Background(), newRecord(slog.LevelInfo, "keep me", "k", "v")); err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	seen := down.seen()
	if len(seen) != 1 {
		t.Fatalf("downstream saw %d records, want 1", len(seen))
	}
	if got := attrsOf(seen[0]); got["k"] != "v" {
		t.Fatalf("attr k = %q, want %q (record should be unchanged)", got["k"], "v")
	}
}

func Test_FilteredLogger_Handle_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		filter  Filter
		attrIn  string
		wantOut string
	}{
		{
			name:    "truncates with ellipsis past the limit",
			filter:  Shorten("body").Limit(10),
			attrIn:  "0123456789ABCDEF",
			wantOut: "0123456...",
		},
		{
			name:    "leaves short values untouched",
			filter:  Shorten("body").Limit(10),
			attrIn:  "short",
			wantOut: "short",
		},
		{
			name:    "default limit of 100 when only keys are set",
			filter:  Shorten("body"),
			attrIn:  "tiny",
			wantOut: "tiny",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			down := newRecHandler(LevelTrace)
			fl := NewFilteredLogger(down, tt.filter)

			rec := newRecord(slog.LevelInfo, "msg", "body", tt.attrIn, "keep", "untouched")
			if err := fl.Handle(context.Background(), rec); err != nil {
				t.Fatalf("Handle() returned error: %v", err)
			}

			seen := down.seen()
			if len(seen) != 1 {
				t.Fatalf("downstream saw %d records, want 1", len(seen))
			}
			got := attrsOf(seen[0])
			if got["body"] != tt.wantOut {
				t.Fatalf("body = %q, want %q", got["body"], tt.wantOut)
			}
			if got["keep"] != "untouched" {
				t.Fatalf("keep = %q, want it left untouched", got["keep"])
			}
		})
	}
}

func Test_FilteredLogger_Handle_ShortenNoDuplicateAttrs(t *testing.T) {
	down := newRecHandler(LevelTrace)
	fl := NewFilteredLogger(down, Shorten("body").Limit(5))

	rec := newRecord(slog.LevelInfo, "msg", "body", "0123456789")
	if err := fl.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	seen := down.seen()
	count := 0
	seen[0].Attrs(func(a slog.Attr) bool {
		if a.Key == "body" {
			count++
		}
		return true
	})
	if count != 1 {
		t.Fatalf("body appeared %d times, want exactly 1 (no duplication)", count)
	}
}

func Test_shortenMessage(t *testing.T) {
	tests := []struct {
		name  string
		msg   string
		limit int
		want  string
	}{
		{"under limit", "abc", 10, "abc"},
		{"equal to limit", "abcde", 5, "abcde"},
		{"over limit with ellipsis", "abcdefgh", 6, "abc..."},
		{"tiny limit no room for ellipsis", "abcdef", 3, "abc"},
		{"zero limit", "abcdef", 0, ""},
		{"negative limit does not panic", "abcdef", -5, ""},
		{"negative limit on empty string", "", -5, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shortenMessage(tt.msg, tt.limit); got != tt.want {
				t.Fatalf("shortenMessage(%q, %d) = %q, want %q", tt.msg, tt.limit, got, tt.want)
			}
		})
	}
}

func Test_FilteredLogger_Handle_ShortenNegativeLimit(t *testing.T) {
	down := newRecHandler(LevelTrace)
	fl := NewFilteredLogger(down, Shorten("body").Limit(-5))

	rec := newRecord(slog.LevelInfo, "msg", "body", "0123456789")
	if err := fl.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	got := attrsOf(down.seen()[0])
	if got["body"] != "" {
		t.Fatalf("body = %q, want %q (negative limit shortens to empty, not panic)", got["body"], "")
	}
}

// Test_FilteredLogger_ConcurrentWithAndAddFilter exercises the lock around the
// filter slice: WithAttrs/WithGroup read it while AddFilter mutates it. Run with
// -race to catch an unguarded read.
func Test_FilteredLogger_ConcurrentWithAndAddFilter(t *testing.T) {
	fl := NewFilteredLogger(noopHandler{})

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				fl.AddFilter(Deny().Message("x"))
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = fl.WithAttrs([]slog.Attr{slog.String("k", "v")})
				_ = fl.WithGroup("g")
			}
		}()
	}
	wg.Wait()
}

func Test_FilteredLogger_AddAndSetFilters(t *testing.T) {
	down := newRecHandler(LevelTrace)
	fl := NewFilteredLogger(down)

	// no filters: record passes through.
	_ = fl.Handle(context.Background(), newRecord(slog.LevelInfo, "one"))

	fl.AddFilter(Deny().Message("two"))
	_ = fl.Handle(context.Background(), newRecord(slog.LevelInfo, "two"))

	fl.SetFilters([]Filter{Deny().Message("three")})
	_ = fl.Handle(context.Background(), newRecord(slog.LevelInfo, "two")) // no longer denied
	_ = fl.Handle(context.Background(), newRecord(slog.LevelInfo, "three"))

	var msgs []string
	for _, r := range down.seen() {
		msgs = append(msgs, r.Message)
	}
	want := []string{"one", "two"}
	if len(msgs) != len(want) {
		t.Fatalf("downstream messages = %v, want %v", msgs, want)
	}
	for i := range want {
		if msgs[i] != want[i] {
			t.Fatalf("downstream messages = %v, want %v", msgs, want)
		}
	}
}
