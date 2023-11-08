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
	level    slog.Level
}

func NewLogHandler(recorder CanRecord, passThroughHandler slog.Handler, level slog.Level) slog.Handler {
	return &SlogHandler{
		recorder: recorder,
		handler:  passThroughHandler,
		level:    level,
	}
}

// Enabled implements slog.Handler.
func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle implements slog.Handler.
func (h *SlogHandler) Handle(ctx context.Context, rec slog.Record) error {
	if rec.Level >= h.level {
		g := h.recorder.Group(rec.Message)
		rec.Attrs(func(a slog.Attr) bool {
			g.Record(rec.Level, a.Key, a.Value.Any())
			return true
		})
	}
	if h.handler.Enabled(ctx, rec.Level) {
		return h.handler.Handle(ctx, rec)
	}
	return nil
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
	return h
}
