package nanny

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"
)

type recorder struct {
	mutex       sync.RWMutex
	events      []Event
	maxEvents   int
	groupMarker string
	stats       *recordingStats
}

type recordingStats struct {
	Started time.Time
	Count   int64
}

type RecorderOption func(*recorder)

func WithMaxEvents(maxEvents int) RecorderOption {
	return func(r *recorder) {
		r.maxEvents = maxEvents
	}
}

func WithGroupMarker(marker string) RecorderOption {
	return func(r *recorder) {
		r.groupMarker = marker
	}
}

func NewRecorder(opts ...RecorderOption) *recorder {
	r := &recorder{
		mutex:       sync.RWMutex{},
		events:      []Event{},
		maxEvents:   100,
		groupMarker: "func",
		stats: &recordingStats{
			Started: time.Now(),
			Count:   0,
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *recorder) Record(level slog.Level, group, message string, attrs map[string]any) {
	ev := Event{
		Time:    time.Now(),
		Level:   level,
		Group:   group,
		Message: message,
		Attrs:   snapshotAttrs(attrs),
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.events = append(r.events, ev)
	r.stats.Count++
	// remove old events
	if len(r.events) > r.maxEvents {
		r.events = r.events[1:]
	}
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
	for _, eg := range r.buildGroups() {
		for _, ev := range eg.events {
			lr := slog.NewRecord(ev.Time, ev.Level, ev.Message, 0)
			if ev.Group != "" {
				lr.AddAttrs(slog.Any("nanny.group", ev.Group))
			}
			for k, v := range ev.Attrs {
				lr.AddAttrs(slog.Any(k, v))
			}
			th.Handle(context.Background(), lr)
		}
	}
}

type eventGroup struct {
	name   string
	events []Event
}

// order of events in group are preserved, groups are also in order
func (r *recorder) buildGroups() []eventGroup {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	groups := []eventGroup{{}} // for the no group
	for _, each := range r.events {
		// lookup group
		found := false
		for g, other := range groups {
			if other.name == each.Group {
				groups[g].events = append(groups[g].events, each)
				found = true
				break
			}
		}
		if !found {
			groups = append(groups, eventGroup{
				name:   each.Group,
				events: []Event{each},
			})
		}
	}
	return groups
}

func snapshotAttrs(attrs map[string]any) map[string]any {
	out := make(map[string]any, len(attrs))
	data, err := json.Marshal(attrs)
	if err != nil {
		out["error"] = err.Error()
		return out
	}
	err = json.Unmarshal(data, &out)
	if err != nil {
		out["error"] = err.Error()
		return out
	}
	return out
}
