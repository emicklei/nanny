package nanny

import (
	"context"
	"log/slog"
	"sync"
)

type SlogHandler struct {
	recorder CanRecord
	mutex    sync.Mutex
	attrs    []slog.Attr
	handler  slog.Handler
}

func NewLogHandler(recorder CanRecord, passThroughHandler slog.Handler) slog.Handler {
	return &SlogHandler{
		recorder: recorder,
		handler:  passThroughHandler,
	}
}

// Enabled implements slog.Handler.
func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (h *SlogHandler) Handle(ctx context.Context, rec slog.Record) error {
	g := h.recorder.Group(rec.Message)
	rec.Attrs(func(a slog.Attr) bool {
		g.Record(a.Key, a.Value.Any())
		return true
	})
	return h.handler.Handle(ctx, rec)
}

// WithAttrs implements slog.Handler.
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.attrs = append(h.attrs, attrs...)
	h.handler = h.handler.WithAttrs(attrs)
	return h
}

// WithGroup implements slog.Handler.
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	return h.handler.WithGroup(name)
}
