package nanny

import (
	"context"
	"net/http"
)

type RecordingHandler struct {
	recorder *recorder
	handler  http.Handler
}

var recorder_key = struct{}{}

func (h *RecordingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), recorder_key, h.recorder)
	h.handler.ServeHTTP(w, r.WithContext(ctx))
}

func NewRecordingHandler(h http.Handler, r *recorder) *RecordingHandler {
	return &RecordingHandler{
		recorder: r,
		handler:  h,
	}
}

func RecorderFromContext(ctx context.Context) *recorder {
	return ctx.Value(recorder_key).(*recorder)
}
