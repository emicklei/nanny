package nanny

type EventGroup struct {
	recorder *recorder
	Group    string
}

func (g *EventGroup) Record(name string, value any) *EventGroup {
	g.recorder.record(g.Group, name, value)
	return g
}
