package nanny

import "log/slog"

type EventGroup struct {
	recorder  *recorder
	groupName string
}

func (g *EventGroup) Record(level slog.Level, name string, value any) CanRecord {
	g.recorder.record(level, g.groupName, name, value)
	return g
}

func (g *EventGroup) Debug(name string, value any) CanRecord {
	return g.Record(slog.LevelDebug, name, value)
}

func (g *EventGroup) Info(name string, value any) CanRecord {
	return g.Record(slog.LevelInfo, name, value)
}

func (g *EventGroup) Group(name string) CanRecord {
	return &EventGroup{
		recorder:  g.recorder,
		groupName: g.groupName + "." + name,
	}
}
