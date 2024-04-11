package interval

import (
	"time"
)

// An Interval represents a time interval with a start and end time.
// The start time is inclusive and the end time is exclusive, represented as [start, end).
type Interval struct {
	start time.Time
	end   time.Time
}

// NewInterval creates and returns a new Interval with the specified start and end times.
func NewInterval(start, end time.Time) *Interval {
	return &Interval{
		start: start,
		end:   end,
	}
}

// Start returns the start time of the interval.
func (i *Interval) Start() time.Time {
	return i.start
}

// End returns the end time of the interval.
func (i *Interval) End() time.Time {
	return i.end
}

// Conflicts checks if the current interval conflicts with the given interval.
// Two intervals conflict if they overlap in time.
func (i *Interval) Conflicts(other *Interval) bool {
	return i.start.Before(other.end) && other.start.Before(i.end)
}
