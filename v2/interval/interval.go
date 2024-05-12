package interval

import (
	"fmt"
	"time"

	"github.com/stevegt/timectl/v2/util"
	// . "github.com/stevegt/goadapt"
)

type Interval struct {
	// Id returns the unique identifier of the interval.  This value
	// should remain constant even if the start, end, or priority of the
	// interval is changed.
	Id    uint64
	Start time.Time
	End   time.Time
	// Priority is the priority of the interval.  Priority zero is the
	// lowest priority, and means that the interval is free.
	Priority float64
	// Payload is the content or event associated with the interval.
	Payload any
}

// NewInterval creates and returns a new Interval with the specified start and end times.
func NewInterval(id uint64, start, end time.Time, priority float64) *Interval {
	if end.Sub(start) <= 0 {
		return nil
	}
	return &Interval{
		Id:       id,
		Start:    start,
		End:      end,
		Priority: priority,
	}
}

// String returns a string representation of the interval.
func (i *Interval) String() string {
	startStr := i.Start.Format(time.RFC3339)
	endStr := i.End.Format(time.RFC3339)
	return fmt.Sprintf("%v %v - %v %v", i.Id, startStr, endStr, i.Priority)
}

// Conflicts checks if the current interval conflicts with the given interval.
// Two intervals conflict if they overlap in time.  If the includeFree
// parameter is true, then a conflict is also detected if either interval
// is free (priority 0).  If the includeFree parameter is false, then
// a conflict is only detected if both intervals are busy (priority > 0).
func (i *Interval) Conflicts(other *Interval, includeFree bool) bool {
	if !includeFree {
		if i.Priority == 0 || other.Priority == 0 {
			return false
		}
	}
	if i.Start.Before(other.End) && i.End.After(other.Start) {
		return true
	}
	if other.Start.Before(i.End) && other.End.After(i.Start) {
		return true
	}
	return false
}

// Equal checks if the current interval is equal to the given interval.
// Two intervals are equal if their start and end times are the same.
func (i *Interval) Equal(other *Interval) bool {
	// this is too strict
	// return i.Start.Equal(other.Start) && i.End.Equal(other.End)
	// XXX tolerance should be an argument
	tolerance := time.Second
	startDiff := i.Start.Sub(other.Start)
	endDiff := i.End.Sub(other.End)

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
func (i *Interval) Wraps(other *Interval) bool {
	if other.Start.Before(i.Start) {
		return false
	}
	if other.End.After(i.End) {
		return false
	}
	return true
}

// Duration returns the duration of the interval.
func (i *Interval) Duration() time.Duration {
	return i.End.Sub(i.Start)
}

// Busy returns true if the interval is not free.  The interval is
// free if the priority is zero.
func (i *Interval) Busy() bool {
	if i.Priority == 0 {
		return false
	}
	return true
}

/*
// Punch creates one to three new intervals by punching a hole in the
// current interval.  The current interval must not be busy and must
// completely contain the hole interval.  The hole interval must be
// busy. Punch does not modify the current interval.
func (i *IntervalBase) Punch(hole Interval) (intervals []Interval) {
	if i.Busy() || !i.Wraps(hole) || !hole.Busy() {
		return nil
	}
	if i.Start.Before(hole.Start) {
		intervals = append(intervals, NewInterval(i.Start, hole.Start, 0))
	}
	intervals = append(intervals, hole)
	if hole.End.Before(i.End) {
		intervals = append(intervals, NewInterval(hole.End, i.End, 0))
	}
	return intervals
}
*/

/*
// Intersection returns an interval that is the intersection of two
// intervals.  The intersection is the interval that overlaps both
// intervals.
func (i *IntervalBase) Intersection(other Interval) Interval {
	start := util.MaxTime(i.Start, other.Start)
	end := util.MinTime(i.End, other.End)
	if start.Before(end) {
		return NewInterval(start, end, 0)
	}
	return nil
}
*/

/*
// Clone returns a deep copy of the interval.
func (i *IntervalBase) Clone() Interval {
	return NewInterval(i.Start, i.End, i.Priority)
}
*/

// Overlaps returns true if the current interval intersects with the given interval.
func (i *Interval) Overlaps(other *Interval) bool {
	if i.Start.Before(other.End) && i.End.After(other.Start) {
		return true
	}
	if other.Start.Before(i.End) && other.End.After(i.Start) {
		return true
	}
	return false
}

// OverlapsRange returns true if the current interval intersects with the given range.
func (i *Interval) OverlapsRange(start, end time.Time) bool {
	if i.Start.Before(end) && i.End.After(start) {
		return true
	}
	return false
}

// OverlapDuration returns the duration of the overlap between the
// current interval and the given range.
func (i *Interval) OverlapDuration(start, end time.Time) time.Duration {
	maxStart := util.MaxTime(i.Start, start)
	minEnd := util.MinTime(i.End, end)
	duration := minEnd.Sub(maxStart)
	if duration < 0 {
		return 0
	}
	return duration
}

// ContainsTime returns true if the given time is after the start time and
// before the end time of the interval.
func (i *Interval) ContainsTime(t time.Time) bool {
	return t.After(i.Start) && t.Before(i.End)
}

// IsBeforeTime returns true if the end time of the interval is on or before
// the given time.
func (i *Interval) IsBeforeTime(t time.Time) bool {
	return !i.End.After(t)
}

// IsAfterTime returns true if the start time of the interval is on or after
// the given time.
func (i *Interval) IsAfterTime(t time.Time) bool {
	return !i.Start.Before(t)
}
