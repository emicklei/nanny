package nanny

import (
	"encoding/json"
	"net/http"
)

type Browser struct {
	recorder *recorder
}

func NewBrowser(rec *recorder) *Browser {
	return &Browser{recorder: rec}
}

func (b *Browser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	do := r.URL.Query().Get("do")
	if "events" == do {
		b.serveEvents(w, r)
		return
	}
	b.serveIndex(w, r)
}

func (b *Browser) serveEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(b.recorder.events)
}
