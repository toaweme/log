package log

import (
	"context"
	"log/slog"
	"testing"
)

func Benchmark_MultiHandler_Handle(b *testing.B) {
	mh := NewMultiHandler(noopHandler{}, noopHandler{})
	ctx := context.Background()
	rec := newRecord(slog.LevelInfo, "request", "user", "bob")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mh.Handle(ctx, rec)
	}
}
