package nanny

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Browser struct {
	recorder *recorder
	options  BrowserOptions
}

type BrowserOptions struct {
	PageSize int
}

func NewBrowser(rec *recorder, opts ...BrowserOptions) *Browser {
	b := &Browser{recorder: rec}
	if len(opts) > 0 {
		b.options = opts[0]
	} else {
		b.options = BrowserOptions{PageSize: 100}
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
	w.Header().Set("x-nanny-version", Version)
	w.Header().Set("x-nanny-page-size", fmt.Sprintf("%d", b.options.PageSize))
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(b.recorder.events)
}
