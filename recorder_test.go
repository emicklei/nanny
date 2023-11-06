package nanny

import (
	"net/http"
	"testing"
	"time"
)

func TestRecorder(t *testing.T) {
	rec := NewRecorder()
	rec.Record("niks", nil)
	rec.Record("now", time.Now())
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	rec.Record("request", req)
	rec.Log()
}
