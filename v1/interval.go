package timectl

import (
	"fmt"
	"time"
	// . "github.com/stevegt/goadapt"
)

// An Interval represents a time interval with a start and end time.
type Interval interface {
	// Start returns the start time of the interval.
	Start() time.Time
	// End returns the end time of the interval.
	End() time.Time
	// Conflicts checks if the current interval conflicts with the given interval.
	Conflicts(other Interval) bool
	// Equal checks if the current interval is equal to the given interval.
	Equal(other Interval) bool
	// Intersection returns an interval that is the intersection of two intervals.
	Intersection(other Interval) Interval
	// Wraps returns true if the current interval completely contains the other interval.
	Wraps(other Interval) bool
	// Duration returns the duration of the interval.
	Duration() time.Duration
	// Busy returns true if the interval is associated with a payload.
	Busy() bool
	// Punch creates one to three new intervals by punching a hole in the current interval.
	Punch(hole Interval) []Interval
	// Priority returns the priority of the interval.  Priority zero
	// is the lowest priority, and means that the interval is free.
	Priority() float64
}

// IntervalBase is the base type for all interval types.
type IntervalBase struct {
	start    time.Time
	end      time.Time
	priority float64
}

// NewInterval creates and returns a new Interval with the specified start and end times.
func NewInterval(start, end time.Time, priority float64) Interval {
	if end.Sub(start) <= 0 {
		return nil
	}
	return &IntervalBase{
		start:    start,
		end:      end,
		priority: priority,
	}
}

// String returns a string representation of the interval.
func (i *IntervalBase) String() string {
	startStr := i.Start().Format(time.RFC3339)
	endStr := i.End().Format(time.RFC3339)
	return fmt.Sprintf("%v - %v %v", startStr, endStr, i.Priority())
}

// Start returns the start time of the interval.
func (i *IntervalBase) Start() time.Time {
	return i.start
}

// End returns the end time of the interval.
func (i *IntervalBase) End() time.Time {
	return i.end
}

// Conflicts checks if the current interval conflicts with the given interval.
// Two intervals conflict if they overlap in time.
func (i *IntervalBase) Conflicts(other Interval) bool {
	if i.Priority() == 0 || other.Priority() == 0 {
		return false
	}
	if i.Start().Before(other.End()) && i.End().After(other.Start()) {
		return true
	}
	if other.Start().Before(i.End()) && other.End().After(i.Start()) {
		return true
	}
	return false
}

// Equal checks if the current interval is equal to the given interval.
// Two intervals are equal if their start and end times are the same.
func (i *IntervalBase) Equal(other Interval) bool {
	// this is too strict
	// return i.Start().Equal(other.Start()) && i.End().Equal(other.End())
	// XXX tolerance should be an argument
	tolerance := time.Second
	startDiff := i.Start().Sub(other.Start())
	endDiff := i.End().Sub(other.End())
	if startDiff < -tolerance || startDiff > tolerance {
		return false
	}
	if endDiff < -tolerance || endDiff > tolerance {
		return false
	}
	return true
}

// Wraps returns true if the current interval completely contains the
// other interval.  In other words, the current interval's start time is
// before or equal to the other interval's start time, and the current
// interval's end time is after or equal to the other interval's end time.
func (i *IntervalBase) Wraps(other Interval) bool {
	if other.Start().Before(i.Start()) {
		return false
	}
	if other.End().After(i.End()) {
		return false
	}
	return true
}

// Duration returns the duration of the interval.
func (i *IntervalBase) Duration() time.Duration {
	return i.End().Sub(i.Start())
}

// Busy returns true if the interval is not free.  The interval is
// free if the priority is zero.
func (i *IntervalBase) Busy() bool {
	if i.Priority() == 0 {
		return false
	}
	return true
}

// Punch creates one to three new intervals by punching a hole in the
// current interval.  The current interval must not be busy and must
// completely contain the hole interval.  The hole interval must be
// busy. Punch does not modify the current interval.
func (i *IntervalBase) Punch(hole Interval) (intervals []Interval) {
	if i.Busy() || !i.Wraps(hole) || !hole.Busy() {
		return nil
	}
	if i.Start().Before(hole.Start()) {
		intervals = append(intervals, NewInterval(i.Start(), hole.Start(), 0))
	}
	intervals = append(intervals, hole)
	if hole.End().Before(i.End()) {
		intervals = append(intervals, NewInterval(hole.End(), i.End(), 0))
	}
	return intervals
}

// Priority returns the priority of the interval.  Priority zero is the
// lowest priority, and means that the interval is free.
func (i *IntervalBase) Priority() float64 {
	return i.priority
}

// Intersection returns an interval that is the intersection of two
// intervals.  The intersection is the interval that overlaps both
// intervals.
func (i *IntervalBase) Intersection(other Interval) Interval {
	start := MaxTime(i.Start(), other.Start())
	end := MinTime(i.End(), other.End())
	if start.Before(end) {
		return NewInterval(start, end, 0)
	}
	return nil
}
