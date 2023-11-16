package nanny

import "log/slog"

type EventGroup struct {
	recorder *recorder
	group    string
}

func (g *EventGroup) Record(level slog.Level, message, name string, value any) CanRecord {
	g.recorder.record(level, g.group, message, name, value)
	return g
}

func (g *EventGroup) Group(name string) CanRecord {
	return &EventGroup{
		recorder: g.recorder,
		group:    g.group + "." + name,
	}
}
