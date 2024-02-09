package nanny

import (
	"context"
	"errors"
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
	slog.Error("test", "g", "h", "err", errors.New("problem"))

	gl := slog.Default().WithGroup("grp")
	gl.Info("test1", "i", "j")
	gl.Info("test2", "k", "l")

	fmt.Println("recorded log:")
	r.log()
}

func TestHandlerOnDefault(t *testing.T) {
	r := NewRecorder()
	h := NewLogHandler(r, slog.Default().Handler(), slog.LevelInfo)
	slog.SetDefault(slog.New(h))
	// TODO check for WARN
	slog.Log(context.TODO(), slog.LevelInfo, "info")
}
