package log

import (
	"context"
	"log/slog"
	"testing"
)

// these benchmarks all call b.ReportAllocs so allocs/op shows up alongside
// ns/op. The hot allocators in the filter path are matchesFilter (builds a
// per-record attr map) and the Shorten action (rebuilds the record).

func Benchmark_shortenMessage(b *testing.B) {
	msg := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = shortenMessage(msg, 20)
	}
}

func Benchmark_FilteredLogger_matchesFilter_Exact(b *testing.B) {
	fl := &FilteredLogger{}
	filter := Allow().Message("request").Attr("user", "bob")
	rec := newRecord(slog.LevelInfo, "request", "user", "bob", "path", "/api/users")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fl.matchesFilter(rec, filter)
	}
}

func Benchmark_FilteredLogger_matchesFilter_Wildcard(b *testing.B) {
	fl := &FilteredLogger{}
	filter := Allow().Attr("path", "/api/*")
	rec := newRecord(slog.LevelInfo, "request", "user", "bob", "path", "/api/users")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fl.matchesFilter(rec, filter)
	}
}

func Benchmark_FilteredLogger_Handle_PassThrough(b *testing.B) {
	fl := NewFilteredLogger(noopHandler{}, Deny().Message("never"))
	ctx := context.Background()
	rec := newRecord(slog.LevelInfo, "request", "user", "bob", "path", "/api/users")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fl.Handle(ctx, rec)
	}
}

func Benchmark_FilteredLogger_Handle_Deny(b *testing.B) {
	fl := NewFilteredLogger(noopHandler{}, Deny().Below(slog.LevelInfo))
	ctx := context.Background()
	rec := newRecord(slog.LevelDebug, "noise", "user", "bob")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fl.Handle(ctx, rec)
	}
}

func Benchmark_FilteredLogger_Handle_Shorten(b *testing.B) {
	fl := NewFilteredLogger(noopHandler{}, Shorten("body").Limit(20))
	ctx := context.Background()
	rec := newRecord(slog.LevelInfo, "response",
		"body", "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"status", "200",
	)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fl.Handle(ctx, rec)
	}
}

func Benchmark_MultiHandler_Handle(b *testing.B) {
	mh := NewMultiHandler(noopHandler{}, noopHandler{})
	ctx := context.Background()
	rec := newRecord(slog.LevelInfo, "request", "user", "bob")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mh.Handle(ctx, rec)
	}
}
