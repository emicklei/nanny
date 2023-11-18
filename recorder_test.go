package nanny

import (
	"log/slog"
	"testing"
)

func TestRecorder(t *testing.T) {
	rec := NewRecorder()
	rec.Record(slog.LevelDebug, "grp", "msg", nil)
	rec.Log()
}

func TestGroupMarker(t *testing.T) {
	rec := NewRecorder(WithGroupMarker("grp"))
	l := slog.New(NewLogHandler(rec, slog.Default().Handler(), slog.LevelDebug))

	g := l.With("func", "TestGroupMarker")
	g.Debug("question", "answer", 42)

	rec.Log()
}
