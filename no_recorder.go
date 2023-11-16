package nanny

import "log/slog"

type NoRecorder struct{}

func (n NoRecorder) Record(slog.Level, string, string, any) CanRecord { return n }
func (n NoRecorder) Group(name string) CanRecord {
	return n
}
