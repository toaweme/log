package log

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// recHandler captures every record it is asked to handle, for assertions. Its
// minimum level is configurable so Enabled behavior can be exercised.
type recHandler struct {
	mu      *sync.Mutex
	records *[]slog.Record
	min     slog.Level
	attrs   []slog.Attr
	groups  []string
}

var _ slog.Handler = (*recHandler)(nil)

func newRecHandler(minLevel slog.Level) *recHandler {
	return &recHandler{
		mu:      &sync.Mutex{},
		records: &[]slog.Record{},
		min:     minLevel,
	}
}

func (h *recHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.min
}

func (h *recHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	*h.records = append(*h.records, r)
	return nil
}

func (h *recHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := *h
	cp.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &cp
}

func (h *recHandler) WithGroup(name string) slog.Handler {
	cp := *h
	cp.groups = append(append([]string{}, h.groups...), name)
	return &cp
}

// fakeHandler records the messages it handled and can be told to fail or to
// gate a minimum level, for MultiHandler tests.
type fakeHandler struct {
	mu        *sync.Mutex
	handled   *[]string
	min       slog.Level
	handleErr error
}

var _ slog.Handler = (*fakeHandler)(nil)

func newFakeHandler(minLevel slog.Level, handleErr error) *fakeHandler {
	return &fakeHandler{
		mu:        &sync.Mutex{},
		handled:   &[]string{},
		min:       minLevel,
		handleErr: handleErr,
	}
}

func (h *fakeHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.min
}

func (h *fakeHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	*h.handled = append(*h.handled, r.Message)
	return h.handleErr
}

func (h *fakeHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *fakeHandler) WithGroup(string) slog.Handler      { return h }

func (h *fakeHandler) messages() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string{}, *h.handled...)
}

// noopHandler discards everything and allocates nothing, so benchmarks measure
// the handler under test rather than the sink.
type noopHandler struct{}

var _ slog.Handler = noopHandler{}

func (noopHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (noopHandler) Handle(context.Context, slog.Record) error { return nil }
func (h noopHandler) WithAttrs([]slog.Attr) slog.Handler      { return h }
func (h noopHandler) WithGroup(string) slog.Handler           { return h }

// newRecord builds a record at level with msg and the given key/value attrs. It
// uses a zero timestamp so records are deterministic.
func newRecord(level slog.Level, msg string, kv ...string) slog.Record {
	r := slog.NewRecord(time.Time{}, level, msg, 0)
	for i := 0; i+1 < len(kv); i += 2 {
		r.AddAttrs(slog.String(kv[i], kv[i+1]))
	}
	return r
}
