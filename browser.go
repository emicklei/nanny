package nanny

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Browser struct {
	recorder *recorder
	pageSize int
}

type BrowserOption func(b *Browser)

func WithPageSize(size int) BrowserOption {
	return func(b *Browser) {
		b.pageSize = size
	}
}

func NewBrowser(rec *recorder, opts ...BrowserOption) *Browser {
	b := &Browser{
		pageSize: 1000,
		recorder: rec}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *Browser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	do := r.URL.Query().Get("do")
	switch do {
	case "events":
		b.serveEvents(w, r)
	case "stop":
		b.recorder.stop()
	case "flush":
		b.recorder.flush()
	case "resume":
		b.recorder.resume()
	default:
		b.serveStaticIndex(w, r)
	}
}

func (b *Browser) serveEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("x-nanny-stats-count", fmt.Sprintf("%d", b.recorder.stats.Count))
	w.Header().Set("x-nanny-stats-started-seconds", fmt.Sprintf("%d", b.recorder.stats.Started.Unix()))
	w.Header().Set("x-nanny-stats-memory-bytes", fmt.Sprintf("%d", b.recorder.computeEventsMemory()))
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(b.recorder.events)
}
