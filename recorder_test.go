package nanny

import (
	"fmt"
	"log/slog"
	"testing"
)

func TestRecorder(t *testing.T) {
	rec := NewRecorder()
	rec.Record(slog.LevelDebug, "grp", "msg", nil)
	rec.Log()
}

func TestGroupMarker(t *testing.T) {
	rec := NewRecorder(WithGroupMarkers("grp"))
	l := slog.New(NewLogHandler(rec, slog.Default().Handler(), slog.LevelDebug))

	g := l.With("func", "TestGroupMarker")
	g.Debug("question", "answer", 42)

	rec.Log()
}

func TestMaxEventGroups(t *testing.T) {
	rec := NewRecorder()
	rec.retentionStrategy = MaxEventGroupsStrategy{MaxEventGroups: 2}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			rec.Record(slog.LevelDebug, fmt.Sprintf("grp%d", i), fmt.Sprintf("msg%d", j), nil)
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
