package nanny

type RetentionStrategy interface {
	// PostRecordedEventBy is called within a locked mutex.
	PostRecordedEventBy(*recorder)
}

type MaxEventsStrategy struct {
	MaxEvents int
}

func (s MaxEventsStrategy) PostRecordedEventBy(r *recorder) {
	if len(r.events) > s.MaxEvents {
		r.events = r.events[1:]
	}
}

type MaxEventGroupsStrategy struct {
	MaxEventGroups int
}

func (s MaxEventGroupsStrategy) PostRecordedEventBy(r *recorder) {
	if len(r.groupSet) > s.MaxEventGroups {
		r.removeOldestEventGroup()
	}
}
