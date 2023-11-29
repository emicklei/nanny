package nanny

import (
	"context"
	"log/slog"
)

type SlogHandler struct {
	recorder  *recorder
	attrs     []slog.Attr
	handler   slog.Handler
	group     string // group name of messages
	attrGroup string
	level     slog.Level
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
		if h.attrGroup != "" {
			// nest it
			collect = map[string]any{
				h.attrGroup: collect,
			}
		}
		h.recorder.Record(h.handler, rec.Level, h.group, rec.Message, collect)
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
	if v := findGroupMarkerValue(attrs, h.recorder.groupMarkers); v != "" {
		gh := NewLogHandler(h.recorder, h.handler.WithAttrs(attrs), h.level)
		gh.attrGroup = h.attrGroup
		ng := v
		if h.group == "" {
			gh.group = ng
		} else {
			gh.group = h.group + "/" + ng
		}
		return gh
	}
	ah := NewLogHandler(h.recorder, h.handler.WithAttrs(attrs), h.level)
	ah.group = h.group
	ah.attrGroup = h.attrGroup
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
	gh.attrGroup = name
	return gh
}

func findGroupMarkerValue(attrs []slog.Attr, markers []string) string {
	for _, marker := range markers {
		for _, a := range attrs {
			if a.Key == marker {
				return a.Value.String() // first come first serve
			}
		}
	}
	return ""
}
