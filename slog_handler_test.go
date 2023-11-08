package nanny

import (
	"log/slog"
	"testing"
)

func TestLogHandler(t *testing.T) {
	r := NewRecorder()
	h := slog.Default().Handler()
	l := slog.New(NewLogHandler(r, h))
	slog.SetDefault(l)

	slog.Info("test", "a", "b")
}
