package nanny

import (
	"log/slog"
	"net/http"
	"time"
)

type Event struct {
	Time    time.Time      `json:"t" `
	Level   slog.Level     `json:"l" `
	Message string         `json:"m" `
	Group   string         `json:"g" ` // group name of events
	Attrs   map[string]any `json:"a" `
}

// SetupDefault wraps the default log handler with a handler that records events.
// Logging all events on error detection is true.
// Maxium number of event groups in memory is 100.
// Installs the browser (page=100) on the default serve mux with path /nanny
func SetupDefault() {
	rec := NewRecorder(
		WithLogEventGroupOnError(true),
		WithMaxEventGroups(100),
		WithGroupMarkers("func"))
	reclog := slog.New(NewLogHandler(rec, slog.Default().Handler(), LevelTrace))
	slog.SetDefault(reclog)
	http.Handle("/nanny", NewBrowser(rec, WithPageSize(100)))
}
