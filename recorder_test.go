package nanny

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
)

func TestRecorder(t *testing.T) {
	rec := NewRecorder()
	rec.Record(slog.Default().Handler(), slog.LevelDebug, "grp", "msg", nil)
	rec.Log()
}

func TestRecorderLogEventGroupOnError(t *testing.T) {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
	rec := NewRecorder(WithLogEventGroupOnError(true))
	rec.Record(h, slog.LevelDebug, "", "ungrouped msg", nil)
	rec.Record(h, slog.LevelDebug, "grp", "msg", nil)
	rec.Record(h, slog.LevelError, "grp", "bummer!", nil)
}

func TestGroupMarker(t *testing.T) {
	rec := NewRecorder(WithGroupMarkers("grp"))
	l := slog.New(NewLogHandler(rec, slog.Default().Handler(), slog.LevelDebug))

	g := l.With("func", "TestGroupMarker")
	g.Debug("question", "answer", 42)

	rec.Log()
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
	rec := NewRecorder()
	rec.retentionStrategy = MaxEventGroupsStrategy{MaxEventGroups: 2}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			rec.Record(slog.Default().Handler(), slog.LevelDebug, fmt.Sprintf("grp%d", i), fmt.Sprintf("msg%d", j), nil)
		}
	}
	if got, want := len(rec.groupSet), 2; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
	grps := rec.buildGroups() // includes the empty group
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
