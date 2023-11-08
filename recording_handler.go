package nanny

import (
	"context"
	"net/http"
)

type RecordingHTTPHandler struct {
	recorder *recorder
	handler  http.Handler
}

var recorder_key = struct{}{}

func (h *RecordingHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), recorder_key, h.recorder)
	h.handler.ServeHTTP(w, r.WithContext(ctx))
}

func NewRecordingHTTPHandler(h http.Handler, r *recorder) *RecordingHTTPHandler {
	return &RecordingHTTPHandler{
		recorder: r,
		handler:  h,
	}
}

func RecorderFromContext(ctx context.Context) CanRecord {
	v := ctx.Value(recorder_key)
	if v, ok := v.(CanRecord); ok {
		return v
	}
	return NoRecorder{}
}
