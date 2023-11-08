package nanny

import (
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestRecorder(t *testing.T) {
	rec := NewRecorder()
	rec.Record(slog.LevelDebug, "niks", nil)
	rec.Record(slog.LevelInfo, "now", time.Now())
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	rec.Record(slog.LevelWarn, "request", req)
	rec.Log()
}
