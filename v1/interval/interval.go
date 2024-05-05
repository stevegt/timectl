package interval

import (
	"fmt"
	"time"

	"github.com/stevegt/timectl/util"
	// . "github.com/stevegt/goadapt"
)

// An Interval represents a time interval with a start and end time.
type Interval interface {
	// Start returns the start time of the interval.
	Start() time.Time
	// End returns the end time of the interval.
	End() time.Time
	// Conflicts checks if the current interval conflicts with the given interval.
	Conflicts(other Interval, includeFree bool) bool
	// Equal checks if the current interval is equal to the given interval.
	Equal(other Interval) bool
	// Intersection returns an interval that is the intersection of two intervals.
	Intersection(other Interval) Interval
	// Wraps returns true if the current interval completely contains the other interval.
	Wraps(other Interval) bool
	// OverlapDuration returns the duration of the overlap between the current interval and the given range.
	OverlapDuration(start, end time.Time) time.Duration
	// Overlaps returns true if the current interval intersects with the given interval.
	Overlaps(other Interval) bool
	// OverlapsRange returns true if the current interval intersects with the given range.
	OverlapsRange(start, end time.Time) bool
	// Duration returns the duration of the interval.
	Duration() time.Duration
	// Busy returns true if the interval is busy.  The interval is busy if the priority is greater than zero.
	Busy() bool
	// Punch creates one to three new intervals by punching a hole in the current interval.
	Punch(hole Interval) []Interval
	// Priority returns the priority of the interval.  Priority zero
	// is the lowest priority, and means that the interval is free.
	Priority() float64
	// SetStart sets the start time of the interval.
	SetStart(time.Time)
	// SetEnd sets the end time of the interval.
	SetEnd(time.Time)
	// SetPriority sets the priority of the interval.
	SetPriority(float64)
	// Clone returns a deep copy of the interval.
	Clone() Interval
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
// Two intervals conflict if they overlap in time.  If the includeFree
// parameter is true, then a conflict is also detected if either interval
// is free (priority 0).  If the includeFree parameter is false, then
// a conflict is only detected if both intervals are busy (priority > 0).
func (i *IntervalBase) Conflicts(other Interval, includeFree bool) bool {
	if !includeFree {
		if i.Priority() == 0 || other.Priority() == 0 {
			return false
		}
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
	start := util.MaxTime(i.Start(), other.Start())
	end := util.MinTime(i.End(), other.End())
	if start.Before(end) {
		return NewInterval(start, end, 0)
	}
	return nil
}

// SetStart sets the start time of the interval.
func (i *IntervalBase) SetStart(start time.Time) {
	i.start = start
}

// SetEnd sets the end time of the interval.
func (i *IntervalBase) SetEnd(end time.Time) {
	i.end = end
}

// SetPriority sets the priority of the interval.
func (i *IntervalBase) SetPriority(priority float64) {
	i.priority = priority
}

// Clone returns a deep copy of the interval.
func (i *IntervalBase) Clone() Interval {
	return NewInterval(i.Start(), i.End(), i.Priority())
}

// Overlaps returns true if the current interval intersects with the given interval.
func (i *IntervalBase) Overlaps(other Interval) bool {
	if i.Start().Before(other.End()) && i.End().After(other.Start()) {
		return true
	}
	if other.Start().Before(i.End()) && other.End().After(i.Start()) {
		return true
	}
	return false
}

// OverlapsRange returns true if the current interval intersects with the given range.
func (i *IntervalBase) OverlapsRange(start, end time.Time) bool {
	if i.Start().Before(end) && i.End().After(start) {
		return true
	}
	return false
}

// OverlapDuration returns the duration of the overlap between the
// current interval and the given range.
func (i *IntervalBase) OverlapDuration(start, end time.Time) time.Duration {
	maxStart := util.MaxTime(i.Start(), start)
	minEnd := util.MinTime(i.End(), end)
	duration := minEnd.Sub(maxStart)
	if duration < 0 {
		return 0
	}
	return duration
}
