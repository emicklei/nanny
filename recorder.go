package nanny

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"
)

var eventPool = sync.Pool{
	New: func() any {
		return new(Event)
	},
}

type recorder struct {
	mutex       sync.RWMutex
	events      []*Event
	options     RecorderOptions
	groupSet    map[string]bool
	stats       *recordingStats
	isRecording bool
}

type recordingStats struct {
	Started time.Time
	Count   int64
}

func NewRecorder(opts ...RecorderOptions) *recorder {
	r := &recorder{
		mutex:    sync.RWMutex{},
		events:   []*Event{},
		groupSet: map[string]bool{},
		stats: &recordingStats{
			Started: time.Now(),
			Count:   0,
		},
		isRecording: true,
	}
	if len(opts) > 0 {
		r.options = opts[0]
	} else {
		r.options = defaultOptions
	}
	return r
}

func (r *recorder) Record(fallback slog.Handler, level slog.Level, group, message string, attrs map[string]any) {
	if !r.isRecording {
		return
	}
	ev := eventPool.Get().(*Event)
	ev.Time = time.Now()
	ev.Level = level
	ev.Group = group
	ev.Message = message
	ev.Attrs = snapshotAttrs(attrs)
	ev.computeMemory()

	r.mutex.Lock()
	r.events = append(r.events, ev)
	r.stats.Count++
	if group != "" {
		// update count cache
		r.groupSet[group] = true
	}
	r.options.postRecordedEventBy(r)
	r.mutex.Unlock()

	if level >= slog.LevelError && r.options.LogEventGroupOnError {
		r.logEventGroup(fallback, group)
	}
}

func (r *recorder) logEventGroup(handler slog.Handler, group string) {
	// make copy so new Record calls are not blocked
	list := r.snapshotEvents()
	ctx := context.Background()
	for _, ev := range list {
		if ev.Group != group {
			continue
		}
		if handler.Enabled(ctx, ev.Level) {
			// already logged
			continue
		}
		lr := slog.NewRecord(ev.Time, ev.Level, ev.Message, 0)
		for k, v := range ev.Attrs {
			lr.AddAttrs(slog.Any(k, v))
		}
		handler.Handle(ctx, lr)
	}
}

func (r *recorder) snapshotEvents() []*Event {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	list := make([]*Event, len(r.events))
	copy(list, r.events)
	return list
}

func (r *recorder) stop() {
	r.isRecording = false
}

func (r *recorder) resume() {
	r.isRecording = true
}

func (r *recorder) flush() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, ev := range r.events {
		eventPool.Put(ev)
	}
	r.events = []*Event{}
}

func (r *recorder) clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, ev := range r.events {
		eventPool.Put(ev)
	}
	// clear cache
	r.groupSet = map[string]bool{}
	r.events = []*Event{}
}

func (r *recorder) computeEventsMemory() (size int64) {
	for _, each := range r.events {
		size += int64(each.memorySize)
	}
	return
}

// pre: mutex lock is active
func (r *recorder) removeFirstEvent() {
	if len(r.events) == 0 {
		return
	}
	oldest := r.events[0]
	r.events = r.events[1:]
	eventPool.Put(oldest)
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
	remaining := []*Event{}
	for _, each := range r.events {
		if each.Group == target {
			eventPool.Put(each)
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
		out["marshal.error"] = err.Error()
		return out
	}
	// always succeeds
	json.Unmarshal(data, &out)
	return out
}
