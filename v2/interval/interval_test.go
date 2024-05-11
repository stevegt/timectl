package interval

import (
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
)

// This package is an implementation of interval trees, optimized for
// use with time intervals for calendaring and scheduling
// applications.

func TestInterval(t *testing.T) {
	// Interval is a type that represents time interval with a start
	// and end time.  The start time is inclusive and the end time is
	// exclusive.  The interval is represented as [start, end).
	start, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	interval := NewInterval(1, start, end, 0)
	Tassert(t, interval.Start() == start, "start time: expected %v, got %v", start, interval.Start())
	Tassert(t, interval.End() == end, "end time: expected %v, got %v", end, interval.End())
}

// TestConflict tests two intervals for conflict.  Two intervals conflict
// if they overlap in time.
func TestConflict(t *testing.T) {
	// Two intervals conflict if they overlap in time.  The intervals
	// [start1, end1) and [start2, end2) conflict if either start1 is
	// between start2 and end2 or end1 is between start2 and end2.
	start1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	interval1 := NewInterval(1, start1, end1, 1)
	start2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:30:00")
	Ck(err)
	end2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:30:00")
	Ck(err)
	interval2 := NewInterval(2, start2, end2, 1)
	Tassert(t, interval1.Conflicts(interval2, false), "expected conflict, got no conflict")
	Tassert(t, interval2.Conflicts(interval1, false), "expected conflict, got no conflict")

	start3, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T09:00:00")
	Ck(err)
	end3, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:30:00")
	interval3 := NewInterval(3, start3, end3, 1)
	Tassert(t, interval1.Conflicts(interval3, false), "expected conflict, got no conflict")
	Tassert(t, interval3.Conflicts(interval1, false), "expected conflict, got no conflict")

	// check identical intervals
	interval3b := NewInterval(4, start3, end3, 1)
	Tassert(t, interval3.Conflicts(interval3b, false), "expected conflict, got no conflict")
	Tassert(t, interval3b.Conflicts(interval3, false), "expected conflict, got no conflict")
}

// TestNoConflict tests two intervals for no conflict.  Two intervals do
// not conflict if they do not overlap in time.
func TestNoConflict(t *testing.T) {
	// Two intervals do not conflict if they do not overlap in time.
	start1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	interval1 := NewInterval(1, start1, end1, 1)
	start2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	end2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T12:00:00")
	Ck(err)
	interval2 := NewInterval(2, start2, end2, 1)
	Tassert(t, !interval1.Conflicts(interval2, false), "expected no conflict, got conflict")
	Tassert(t, !interval2.Conflicts(interval1, false), "expected no conflict, got conflict")

	// Two intervals do not conflict if one is a free slot.
	interval3 := NewInterval(3, start1, end1, 0)
	Tassert(t, !interval1.Conflicts(interval3, false), "expected no conflict, got conflict")
}

// test free conflict
func TestFreeConflict(t *testing.T) {
	// Two intervals conflict if they overlap in time and includeFree is
	// true.  The intervals [start1, end1) and [start2, end2) conflict if
	// either start1 is between start2 and end2 or end1 is between start2
	// and end2.
	start1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	interval1 := NewInterval(1, start1, end1, 0)
	start2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:30:00")
	Ck(err)
	end2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:30:00")
	Ck(err)
	interval2 := NewInterval(2, start2, end2, 1)
	Tassert(t, interval1.Conflicts(interval2, true), "expected conflict, got no conflict")
	Tassert(t, interval2.Conflicts(interval1, true), "expected conflict, got no conflict")

	start3, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T09:00:00")
	Ck(err)
	end3, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:30:00")
	interval3 := NewInterval(3, start3, end3, 1)
	Tassert(t, interval1.Conflicts(interval3, true), "expected conflict, got no conflict")
	Tassert(t, interval3.Conflicts(interval1, true), "expected conflict, got no conflict")

	// check identical intervals
	interval3b := NewInterval(4, start3, end3, 1)
	Tassert(t, interval3.Conflicts(interval3b, true), "expected conflict, got no conflict")
	Tassert(t, interval3b.Conflicts(interval3, true), "expected conflict, got no conflict")
}

// TestEqual tests two intervals for equality.  Two intervals are equal
// if their start and end times are equal.
func TestEqual(t *testing.T) {
	// Two intervals are equal if their start and end times are equal.
	start1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	interval1 := NewInterval(1, start1, end1, 1)
	interval1a := NewInterval(2, start1, end1, 1)
	Tassert(t, interval1.Equal(interval1a), "expected equal, got not equal")
	Tassert(t, interval1a.Equal(interval1), "expected equal, got not equal")

	start2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:01:00")
	Ck(err)
	interval2 := NewInterval(3, start2, end1, 1)
	Tassert(t, !interval1.Equal(interval2), "expected not equal, got equal")

	end2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:50:00")
	Ck(err)
	interval3 := NewInterval(4, start1, end2, 1)
	Tassert(t, !interval1.Equal(interval3), "expected not equal, got equal")
}

func TestWraps(t *testing.T) {
	t1000, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	t1100, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)

	t1001, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:01:00")
	Ck(err)
	t1050, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:50:00")
	Ck(err)

	interval1 := NewInterval(1, t1000, t1100, 1)
	interval1a := NewInterval(2, t1000, t1100, 1)
	Tassert(t, interval1.Wraps(interval1a), "expected interval1 to wrap interval1a")

	interval2 := NewInterval(3, t1000, t1050, 1)
	Tassert(t, !interval2.Wraps(interval1), "expected interval2 to not wrap interval1")

	interval3 := NewInterval(4, t1001, t1100, 1)
	Tassert(t, !interval3.Wraps(interval1), "expected interval3 to not wrap interval1")

	interval4 := NewInterval(5, t1001, t1050, 1)
	Tassert(t, !interval4.Wraps(interval1), "expected interval4 to not wrap interval1")
}

/*
// Intersection returns an interval that is the intersection of two
// intervals.  The intersection is the interval that overlaps both
// intervals.
func TestIntersection(t *testing.T) {
	t1000, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	t1100, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	t1030, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:30:00")
	Ck(err)
	t1130, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:30:00")
	Ck(err)
	t1200, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T12:00:00")

	i1000_1100 := NewInterval(1, t1000, t1100, 1)
	i1030_1130 := NewInterval(2, t1030, t1130, 1)
	i1030_1100 := i1000_1100.Intersection(i1030_1130)
	Tassert(t, i1030_1100.Start() == t1030, "expected start time %v, got %v", t1030, i1030_1100.Start())
	Tassert(t, i1030_1100.End() == t1100, "expected end time %v, got %v", t1100, i1030_1100.End())

	i1130_1200 := NewInterval(3, t1130, t1200, 1)
	iNil := i1000_1100.Intersection(i1130_1200)
	Tassert(t, iNil == nil, "expected nil, got %v", iNil)

}
*/
