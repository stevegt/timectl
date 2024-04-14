package interval

import (
	"fmt"
	"time"
)

// An Interval represents a time interval with a start and end time.
// The start time is inclusive and the end time is exclusive, represented as [start, end).
type Interval struct {
	start time.Time
	end   time.Time
	// event is the event that the interval is associated with.  If
	// event is nil, the interval represents a free slot.
	event any
}

// NewInterval creates and returns a new Interval with the specified start and end times.
func NewInterval(start, end time.Time, event any) *Interval {
	if end.Sub(start) <= 0 {
		return nil
	}
	return &Interval{
		start: start,
		end:   end,
		event: event,
	}
}

// String returns a string representation of the interval.
func (i *Interval) String() string {
	return fmt.Sprintf("%v - %v %v", i.start.Format(time.RFC3339), i.end.Format(time.RFC3339), i.event)
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
	if i.event == nil || other.event == nil {
		return false
	}
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

// Busy returns true if the interval is associated with an event.  The
// interval is free if the event is nil or false, or busy otherwise.
func (i *Interval) Busy() bool {
	if i.event == nil {
		return false
	}
	if i.event == false {
		return false
	}
	return true
}

// Punch creates one to three new intervals by punching a hole in the
// current interval.  The current interval must not be busy and must
// completely contain the hole interval.  The hole interval must be
// busy.
func (i *Interval) Punch(hole *Interval) (intervals []*Interval) {
	if i.Busy() || !i.Wraps(hole) || !hole.Busy() {
		return nil
	}
	if i.start.Before(hole.start) {
		intervals = append(intervals, NewInterval(i.start, hole.start, nil))
	}
	intervals = append(intervals, hole)
	if hole.end.Before(i.end) {
		intervals = append(intervals, NewInterval(hole.end, i.end, nil))
	}
	return intervals
}

// Event returns the event associated with the interval.
func (i *Interval) Event() any {
	return i.event
}
