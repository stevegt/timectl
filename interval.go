package interval

import (
	"time"
	// . "github.com/stevegt/goadapt"
)

// An Interval represents a time interval with a start and end time.
// The start time is inclusive and the end time is exclusive, represented as [start, end).
type Interval struct {
	start time.Time
	end   time.Time
}

// NewInterval creates and returns a new Interval with the specified start and end times.
func NewInterval(start, end time.Time) *Interval {
	if end.Sub(start) <= 0 {
		return nil
	}
	return &Interval{
		start: start,
		end:   end,
	}
}

// String returns a string representation of the interval.
func (i *Interval) String() string {
	return i.start.Format(time.RFC3339) + " - " + i.end.Format(time.RFC3339)
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
	// Pf("i = %v, other = %v\n", i, other)
	if i.start.Before(other.end) && i.end.After(other.start) {
		return true
	}
	if other.start.Before(i.end) && other.end.After(i.start) {
		return true
	}
	return false
}

// Equal checks if the current interval is equal to the given interval.
// Two intervals are equal if their start and end times are the same.
func (i *Interval) Equal(other *Interval) bool {
	return i.start.Equal(other.start) && i.end.Equal(other.end)
}

// Wraps returns true if the current interval completely contains the
// other interval.  In other words, the current interval's start time is
// before or equal to the other interval's start time, and the current
// interval's end time is after or equal to the other interval's end time.
func (i *Interval) Wraps(other *Interval) bool {
	if other.start.Before(i.start) {
		return false
	}
	if other.end.After(i.end) {
		return false
	}
	return true
}

// Duration returns the duration of the interval.
func (i *Interval) Duration() time.Duration {
	return i.end.Sub(i.start)
}
