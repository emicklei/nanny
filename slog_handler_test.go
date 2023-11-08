package nanny

import (
	"log/slog"
	"os"
	"testing"
)

func TestLogHandler(t *testing.T) {
	r := NewRecorder()
	// fallback only info
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

	// recorder captures debug
	l := slog.New(NewLogHandler(r, h, slog.LevelDebug))
	slog.SetDefault(l)

	slog.Debug("debug", "c", "d")
	slog.Info("test", "a", "b")

	r.Log()
}
