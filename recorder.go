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
	mutex                sync.RWMutex
	events               []Event
	retentionStrategy    RetentionStrategy
	groupMarkers         []string
	groupSet             map[string]bool
	stats                *recordingStats
	isRecording          bool
	logEventGroupOnError bool
}

type recordingStats struct {
	Started time.Time
	Count   int64
}

type RecorderOption func(*recorder)

func WithMaxEvents(maxEvents int) RecorderOption {
	return func(r *recorder) {
		r.retentionStrategy = MaxEventsStrategy{MaxEvents: maxEvents}
	}
}

func WithMaxEventGroups(maxGroups int) RecorderOption {
	return func(r *recorder) {
		r.retentionStrategy = MaxEventGroupsStrategy{MaxEventGroups: maxGroups}
	}
}

func WithGroupMarkers(markers ...string) RecorderOption {
	return func(r *recorder) {
		r.groupMarkers = markers
	}
}

func WithLogEventGroupOnError(enabled bool) RecorderOption {
	return func(r *recorder) {
		r.logEventGroupOnError = enabled
	}
}

func NewRecorder(opts ...RecorderOption) *recorder {
	r := &recorder{
		mutex:        sync.RWMutex{},
		events:       []Event{},
		groupMarkers: []string{"func"},
		groupSet:     map[string]bool{},
		stats: &recordingStats{
			Started: time.Now(),
			Count:   0,
		},
		isRecording:          true,
		logEventGroupOnError: false,
		retentionStrategy:    MaxEventsStrategy{MaxEvents: 100},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *recorder) Record(fallback slog.Handler, level slog.Level, group, message string, attrs map[string]any) {
	if !r.isRecording {
		return
	}
	ev := Event{
		Time:    time.Now(),
		Level:   level,
		Group:   group,
		Message: message,
		Attrs:   snapshotAttrs(attrs),
	}
	r.mutex.Lock()
	r.events = append(r.events, ev)
	r.stats.Count++
	if group != "" {
		// update count cache
		r.groupSet[group] = true
	}
	r.retentionStrategy.PostRecordedEventBy(r)
	r.mutex.Unlock()

	if level >= slog.LevelError && r.logEventGroupOnError {
		r.logEventGroup(fallback, group)
	}
}

func (r *recorder) logEventGroup(handler slog.Handler, group string) {
	r.mutex.RLock()
	// make copy to new record calls are not blocked
	list := make([]Event, len(r.events))
	copy(list, r.events)
	r.mutex.RUnlock()

	for _, ev := range list {
		if ev.Group != group {
			continue
		}
		lr := slog.NewRecord(ev.Time, ev.Level, ev.Message, 0)
		lr.AddAttrs(slog.Any("nanny.group", ev.Group))
		for k, v := range ev.Attrs {
			lr.AddAttrs(slog.Any(k, v))
		}
		handler.Handle(context.Background(), lr)
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

func (r *recorder) stop() {
	r.isRecording = false
}

func (r *recorder) resume() {
	r.isRecording = true
}

func (r *recorder) flush() {
	r.mutex.Lock()
	defer r.mutex.RLock()
	r.events = []Event{}
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

// pre: mutex lock is active
func (r *recorder) removeOldestEventGroup() {
	// first group detected is the oldest; events are ordered by time.
	target := ""
	for _, each := range r.events {
		if each.Group == "" {
			continue
		}
		target = each.Group
		break
	}
	// any group?
	if target == "" {
		return
	}
	// remove events by copying
	remaining := []Event{}
	for _, each := range r.events {
		if each.Group == target {
			continue
		}
		remaining = append(remaining, each)
	}
	r.events = remaining
	// update cached set
	delete(r.groupSet, target)
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
