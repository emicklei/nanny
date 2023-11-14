package nanny

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
)

func TestLogHandler(t *testing.T) {
	r := NewRecorder()
	// fallback only info
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

	// recorder captures trace,debug,info,warn,error
	l := slog.New(NewLogHandler(r, h, LevelTrace))
	slog.SetDefault(l)

	slog.Log(context.TODO(), LevelTrace, "trace", "e", "f")
	slog.Debug("debug", "c", "d")
	slog.Info("test", "a", "b")
	slog.Error("test", "g", "h")

	gl := slog.Default().WithGroup("grp")
	gl.Info("test", "g", "h")

	fmt.Println("recorded log:")
	r.Log()
}
