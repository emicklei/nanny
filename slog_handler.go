package nanny

import (
	"context"
	"log/slog"
)

type SlogHandler struct {
	recorder CanRecord
	attrs    []slog.Attr
	handler  slog.Handler
	group    string
	level    slog.Level
}

func NewLogHandler(recorder CanRecord, passThroughHandler slog.Handler, level slog.Level) *SlogHandler {
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
		target := h.recorder
		if h.group != "" {
			target = target.Group(h.group)
		}
		rec.Attrs(func(a slog.Attr) bool {
			target.Record(rec.Level, rec.Message, a.Key, a.Value.Any())
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
	ah := NewLogHandler(h.recorder, h.handler.WithAttrs(attrs), h.level)
	ah.attrs = attrs
	return ah
}

// WithGroup implements slog.Handler.
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	gh := NewLogHandler(h.recorder, h.handler, h.level)
	gh.group = name
	if h.group != "" {
		gh.group = h.group + "." + name
	}
	return gh
}
