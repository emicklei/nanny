package nanny

import (
	"log/slog"
	"time"
)

type Event struct {
	Time    time.Time  `json:"t" `
	Level   slog.Level `json:"l" `
	Message string     `json:"m" `
	Group   string     `json:"g" ` // group name of events
	Attrs   any        `json:"a" `
	Keys    []string   `json:"-" `
}
