package nanny

type EventGroup struct {
	recorder  *recorder
	groupName string
}

func (g *EventGroup) Record(name string, value any) CanRecord {
	g.recorder.record(g.groupName, name, value)
	return g
}
func (g *EventGroup) Group(name string) CanRecord {
	return &EventGroup{
		recorder:  g.recorder,
		groupName: g.groupName + "." + name,
	}
}
