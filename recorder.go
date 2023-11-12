package nanny

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"
)

type Event struct {
	Time      time.Time  `json:"t" `
	Level     slog.Level `json:"l" `
	Name      string     `json:"n" `
	Group     string     `json:"g" `
	ValueType string     `json:"r" `
	Value     any        `json:"v" `
}

type recorder struct {
	mutex     sync.RWMutex
	events    []Event
	maxEvents int
}

type Option func(*recorder)

func WithMaxEvents(maxEvents int) Option {
	return func(r *recorder) {
		r.maxEvents = maxEvents
	}
}

type RetentionStrategy interface {
}

type CanRecord interface {
	Record(level slog.Level, name string, value any) CanRecord
	Group(name string) CanRecord
}

func NewRecorder(opts ...Option) *recorder {
	return &recorder{
		mutex:     sync.RWMutex{},
		events:    []Event{},
		maxEvents: 100,
	}
}

func (r *recorder) Group(name string) CanRecord {
	return &EventGroup{
		recorder:  r,
		groupName: name,
	}
}

func (r *recorder) Record(level slog.Level, name string, value any) CanRecord {
	r.record(level, "", name, value)
	return r
}

func (r *recorder) record(level slog.Level, group, name string, value any) {
	ev := Event{
		Time:      time.Now(),
		Level:     level,
		Group:     group,
		Name:      name,
		ValueType: fmt.Sprintf("%T", value),
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if value == nil {
		r.events = append(r.events, ev)
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		ev.Value = fmt.Errorf("%T cannot be marshalled to JSON: %w", value, err)
		r.events = append(r.events, ev)
		return
	}
	doc := map[string]any{}
	err = json.Unmarshal(data, &doc)
	if err != nil {
		ev.Value = value // store as is
		r.events = append(r.events, ev)
		return
	}
	ev.Value = doc
	r.events = append(r.events, ev)
}

// Log outputs all events using the TextHandler
func (r *recorder) Log() {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	list := make([]Event, len(r.events))
	copy(list, r.events)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Time.Before(list[j].Time)
	})
	// do not use default handler because that could be a recording one
	th := slog.NewTextHandler(os.Stdout, nil)
	for _, ev := range list {
		lr := slog.NewRecord(ev.Time, ev.Level, ev.Group, 0)
		lr.AddAttrs(slog.Any(ev.Name, ev.Value))
		th.Handle(context.Background(), lr)
	}
}
