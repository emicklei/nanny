package nanny

import (
	"context"
	"log/slog"
)

var LevelTrace = slog.Level(-8)

func Trace(msg string, args ...any) {
	slog.Log(context.TODO(), LevelTrace, msg, args...)
}

func TraceContext(ctx context.Context, msg string, args ...any) {
	slog.Log(ctx, LevelTrace, msg, args...)
}
