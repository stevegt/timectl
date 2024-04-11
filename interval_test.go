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
	interval := NewInterval(start, end)
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
	interval1 := NewInterval(start1, end1)
	start2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:30:00")
	Ck(err)
	end2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:30:00")
	Ck(err)
	interval2 := NewInterval(start2, end2)
	Tassert(t, interval1.Conflicts(interval2), "expected conflict, got no conflict")

	start3, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T09:00:00")
	Ck(err)
	end3, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:30:00")
	interval3 := NewInterval(start3, end3)
	Tassert(t, interval1.Conflicts(interval3), "expected conflict, got no conflict")
}

// TestNoConflict tests two intervals for no conflict.  Two intervals do
// not conflict if they do not overlap in time.
func TestNoConflict(t *testing.T) {
	// Two intervals do not conflict if they do not overlap in time.
	start1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	interval1 := NewInterval(start1, end1)
	start2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	end2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T12:00:00")
	Ck(err)
	interval2 := NewInterval(start2, end2)
	Tassert(t, !interval1.Conflicts(interval2), "expected no conflict, got conflict")
}
