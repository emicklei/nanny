package nanny

import (
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestRecorder(t *testing.T) {
	rec := NewRecorder()
	rec.Record(slog.LevelDebug, "msg", "niks", nil)
	rec.Record(slog.LevelInfo, "msg", "now", time.Now())
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	rec.Record(slog.LevelWarn, "msg", "request", req)
	rec.Log()
}
