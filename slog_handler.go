package nanny

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
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
	if !canUseHandler(passThroughHandler) {
		slog.Warn("cannot install nanny handler on nil or default slog handler, using simple text handler on info instead", "handler", passThroughHandler)
		passThroughHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return &SlogHandler{
		recorder: recorder,
		handler:  passThroughHandler,
		level:    level,
	}
}

func canUseHandler(handler slog.Handler) bool {
	if handler == nil {
		return false
	}
	typeName := fmt.Sprintf("%T", handler)
	return !strings.HasSuffix(typeName, "*slog.defaultHandler")
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
			val := a.Value.Any()
			// if an error then add its string value and type too
			if err, ok := val.(error); ok {
				collect[a.Key+".String"] = err.Error()
				collect[a.Key+".type"] = fmt.Sprintf("%T", val)
			}
			collect[a.Key] = val
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
	// check for groupin request
	// no nesting of groups because of potential multiple markers
	if h.group == "" {
		if m, v := findGroupMarkerAndValue(attrs, h.recorder.options.GroupMarkers); v != "" {
			gh := NewLogHandler(h.recorder, h.handler.WithAttrs(attrs), h.level)
			gh.attrGroup = h.attrGroup
			copyAttrs := make([]slog.Attr, 0, len(attrs)+len(h.attrs))
			copyAttrs = append(copyAttrs, h.attrs...)
			for _, a := range attrs {
				if a.Key != m {
					copyAttrs = append(copyAttrs, a)
				}
			}
			gh.attrs = copyAttrs
			gh.group = v
			return gh
		}
	}
	ah := NewLogHandler(h.recorder, h.handler.WithAttrs(attrs), h.level)
	ah.group = h.group
	ah.attrGroup = h.attrGroup
	ah.attrs = append(h.attrs, attrs...)
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

func findGroupMarkerAndValue(attrs []slog.Attr, markers []string) (string, string) {
	for _, marker := range markers {
		for _, a := range attrs {
			if a.Key == marker {
				return marker, a.Value.String() // first come first serve
			}
		}
	}
	return "", ""
}
