package nanny

type RetentionStrategy interface {
	// PostRecordedEventBy is called within a locked mutex.
	PostRecordedEventBy(*recorder)
}

type maxEventsStrategy struct {
	maxEvents int
}

func (s maxEventsStrategy) PostRecordedEventBy(r *recorder) {
	if len(r.events) > s.maxEvents {
		r.events = r.events[1:]
	}
}

type maxEventGroupsStrategy struct {
	maxEventGroups int
	maxEvents      int
}

func (s maxEventGroupsStrategy) PostRecordedEventBy(r *recorder) {
	if len(r.groupSet) > s.maxEventGroups {
		r.removeOldestEventGroup()
		return
	}
	// maxEventsStrategy
	if len(r.events) > s.maxEvents {
		r.events = r.events[1:]
	}
}
