package nanny

import (
	"context"
	"log/slog"
)

var LevelTrace = slog.Level(-8)

func Trace(sl *slog.Logger, msg string, args ...any) {
	sl.Log(context.TODO(), LevelTrace, msg, args...)
}

func TraceContext(ctx context.Context, sl *slog.Logger, msg string, args ...any) {
	sl.Log(ctx, LevelTrace, msg, args...)
}
