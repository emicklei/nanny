package nanny

import (
	"log/slog"
	"testing"
	"time"
)

func TestEventSize(t *testing.T) {
	e := Event{
		Time:    time.Now(),
		Level:   slog.LevelDebug,
		Message: "test",
		Group:   "group",
		Attrs: map[string]any{
			"key": []string{"values"},
		},
	}
	e.computeMemory()
	if got, want := e.memorySize, 236; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
}
