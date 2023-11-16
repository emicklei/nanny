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
