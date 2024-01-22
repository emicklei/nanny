package nanny

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestRecorder(t *testing.T) {
	var someAttrs = map[string]any{
		"pi":    3.14159,
		"color": "pink",
	}
	rec := NewRecorder()
	rec.Record(slog.Default().Handler(), slog.LevelDebug, "grp", "msg", someAttrs)
	e := rec.events[0]
	e.Time, _ = time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	if got, want := e.Level, slog.LevelDebug; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
	if got, want := e.Group, "grp"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
	if got, want := e.Attrs["msg"], any(nil); got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
	if got, want := e.memorySize, 262; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
	if got, want := rec.computeEventsMemory(), int64(262); got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
}

func TestRecorderStopResumeFlush(t *testing.T) {
	rec := NewRecorder()
	rec.Record(slog.Default().Handler(), slog.LevelDebug, "grp", "msg", nil)
	rec.stop()
	rec.flush()
	rec.resume()
	rec.Record(slog.Default().Handler(), slog.LevelDebug, "grp", "msg", nil)
	if got, want := len(rec.events), 1; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
	rec.clear()
	if got, want := len(rec.events), 0; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
}

func TestRecorderLogEventGroupOnError(t *testing.T) {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
	rec := NewRecorder()
	rec.Record(h, slog.LevelDebug, "", "ungrouped msg", nil)
	rec.Record(h, slog.LevelDebug, "grp", "msg", nil)
	rec.Record(h, slog.LevelError, "grp", "bummer!", nil)
}

func TestGroupMarker(t *testing.T) {
	rec := NewRecorder(RecorderOptions{GroupMarkers: []string{"grp"}})
	l := slog.New(NewLogHandler(rec, slog.Default().Handler(), slog.LevelDebug))

	g := l.With("func", "TestGroupMarker")
	g.Debug("question", "answer", 42)

	rec.log()
}

func TestRecorderConditions(t *testing.T) {
	rec := NewRecorder()
	rec.Record(slog.Default().Handler(), slog.LevelDebug, "grp", "hello world", map[string]any{
		"grp": map[string]any{
			"key": "value",
		},
		"shoe": 42,
	})
	ev1 := rec.events[0]
	con1 := RecordCondition{
		Name:    "level debug",
		Enabled: true,
		Path:    "level",
		Value:   "debug",
	}
	if !con1.Matches(ev1) {
		t.Errorf("condition %v did not match event %v", con1, ev1)
	}
	con2 := RecordCondition{
		Name:    "message includes",
		Enabled: true,
		Path:    "message",
		Value:   "/.*world.*/",
	}.withRegexp()
	if !con2.Matches(ev1) {
		t.Errorf("condition %v did not match event %v", con2, ev1)
	}
	con3 := RecordCondition{
		Name:    "message exact",
		Enabled: true,
		Path:    "message",
		Value:   "hello world",
	}
	if !con3.Matches(ev1) {
		t.Errorf("condition %v did not match event %v", con2, ev1)
	}
	con4 := RecordCondition{
		Name:    "attr int",
		Enabled: true,
		Path:    "attrs.shoe",
		Value:   "42",
	}
	if !con4.Matches(ev1) {
		t.Errorf("condition %v did not match event %v", con4, ev1)
	}
	con5 := RecordCondition{
		Name:    "attr int",
		Enabled: true,
		Path:    "attrs.grp.key",
		Value:   "value",
	}
	if !con5.Matches(ev1) {
		t.Errorf("condition %v did not match event %v", con5, ev1)
	}
}

func TestMaxEventGroups(t *testing.T) {
	rec := NewRecorder(RecorderOptions{
		MaxEvents:      1000,
		MaxEventGroups: 2,
	})
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			rec.Record(slog.Default().Handler(), slog.LevelDebug, fmt.Sprintf("grp%d", i), fmt.Sprintf("msg%d", j), nil)
		}
	}
	if got, want := len(rec.groupSet), 2; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
	grps := rec.buildGroups(rec.events) // includes the empty group
	if got, want := len(grps), 3; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
	if got, want := grps[1].name, "grp3"; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
	if got, want := grps[2].name, "grp4"; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
}

func TestSnapshotAttrs(t *testing.T) {
	attrs := map[string]any{
		"invalid": TestSnapshotAttrs,
	}
	m := snapshotAttrs(attrs)
	k := m["marshal.error"]
	if k == nil {
		t.Errorf("missing marshal.error")
	}
}

func TestRemoveOldestEventGroup(t *testing.T) {
	rec := NewRecorder()
	rec.Record(slog.Default().Handler(), slog.LevelDebug, "", "hello", nil)
	rec.Record(slog.Default().Handler(), slog.LevelDebug, "say", "hello", nil)
	rec.removeOldestEventGroup()
	if got, want := len(rec.events), 1; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}

	rec.Record(slog.Default().Handler(), slog.LevelDebug, "grp", "world", nil)
	rec.Record(slog.Default().Handler(), slog.LevelDebug, "", "world", nil)
	rec.removeOldestEventGroup()
	if got, want := len(rec.events), 2; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
}
func TestStop(t *testing.T) {
	rec := NewRecorder()
	rec.stop()
	rec.Record(slog.Default().Handler(), slog.LevelDebug, "", "hello", nil)
	if got, want := len(rec.events), 0; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}

}
