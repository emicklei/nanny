package nanny

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

type Event struct {
	Time      time.Time `json:"t,omitempty" `
	Name      string    `json:"n,omitempty" `
	Group     string    `json:"g,omitempty" `
	ValueType string    `json:"_,omitempty" `
	Value     any       `json:"v"`
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
	Record(name string, value any) CanRecord
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

func (r *recorder) Record(name string, value any) CanRecord {
	r.record("", name, value)
	return r
}

func (r *recorder) record(group, name string, value any) {
	ev := Event{
		Time:      time.Now(),
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

func (r *recorder) Log() {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	list := make([]Event, len(r.events))
	copy(list, r.events)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Time.Before(list[j].Time)
	})
	for _, ev := range list {
		lr := slog.NewRecord(ev.Time, slog.LevelInfo, ev.Name, 0)
		lr.AddAttrs(slog.Any("value", ev.Value))
		slog.Default().Handler().Handle(context.Background(), lr)
	}
}
