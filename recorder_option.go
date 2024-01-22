package nanny

import (
	"log/slog"
	"net/http"

	"github.com/DmitriyVTitov/size"
)

// RecorderOptions holds all configuration options for keeping a window of events.
// The maxima are handled in order of precedence: MaxEventsMemoryBytes->MaxEventGroups->MaxEvents.
type RecorderOptions struct {
	MaxEventsMemoryBytes int64    // zero means no limit
	MaxEventGroups       int      // zero means no limit
	MaxEvents            int      // zero means no limit
	GroupMarkers         []string // one or more attribute names for grouping events
	LogEventGroupOnError bool     // if an Error event is recorded then all leading debug and trace events in the same group are logger first
}

var defaultOptions = RecorderOptions{
	MaxEvents:            1000,
	GroupMarkers:         []string{""},
	LogEventGroupOnError: true,
}

// pre: mutex lock is active
func (o RecorderOptions) postRecordedEventBy(r *recorder) {
	if o.MaxEventsMemoryBytes > 0 {
		mem := r.computeEventsMemory()
		for mem > o.MaxEventsMemoryBytes {
			r.removeOldestEventGroup()
			// abort if no more events
			if len(r.events) == 0 {
				return
			}
			newMem := r.computeEventsMemory()
			if newMem == mem {
				// no clean up, remove last event, which always saves memory
				r.removeFirstEvent()
				mem = r.computeEventsMemory()
			}
		}
		return
	}
	if o.MaxEventGroups > 0 {
		if len(r.groupSet) > o.MaxEventGroups {
			r.removeOldestEventGroup()
			return
		}
	}
	if o.MaxEvents > 0 {
		if len(r.events) > o.MaxEvents {
			r.removeFirstEvent()
		}
		return
	}
}

// SetupDefault wraps the default log handler with a handler that records events.
// Logging all events on error detection is true.
// Maxium number of event groups in memory is 100.
// Installs the browser (page=100) on the default serve mux with path /nanny
func SetupDefault() {
	rec := NewRecorder(defaultOptions)
	reclog := slog.New(NewLogHandler(rec, slog.Default().Handler(), LevelTrace))
	slog.SetDefault(reclog)
	http.Handle("/nanny", NewBrowser(rec, WithPageSize(100)))
}

func (e *Event) computeMemory() {
	e.memorySize = size.Of(*e)
}
