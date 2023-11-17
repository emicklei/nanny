package nanny

import (
	"context"
	"log/slog"
)

type SlogHandler struct {
	recorder *recorder
	attrs    []slog.Attr
	handler  slog.Handler
	group    string // group name of messages
	level    slog.Level
}

func NewLogHandler(recorder *recorder, passThroughHandler slog.Handler, level slog.Level) *SlogHandler {
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
		collect := map[string]any{}
		for _, a := range h.attrs {
			collect[a.Key] = a.Value.Any()
		}
		rec.Attrs(func(a slog.Attr) bool {
			collect[a.Key] = a.Value.Any()
			return true
		})
		h.recorder.Record(rec.Level, h.group, rec.Message, collect)
	}
	if h.handler.Enabled(ctx, rec.Level) {
		return h.handler.Handle(ctx, rec)
	}
	return nil
}

// WithAttrs implements slog.Handler.
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	// todo: only one key with groupmarker
	if len(attrs) == 1 && attrs[0].Value.String() == h.recorder.groupMarker {
		gh := NewLogHandler(h.recorder, h.handler, h.level)
		gh.group = attrs[0].Key
		return gh
	}
	ah := NewLogHandler(h.recorder, h.handler.WithAttrs(attrs), h.level)
	ah.attrs = attrs
	return ah
}

// WithGroup implements slog.Handler.
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	gh := NewLogHandler(h.recorder, h.handler.WithGroup(name), h.level)
	gh.group = h.group // event group, not attr group
	return gh
}
