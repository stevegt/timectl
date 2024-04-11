package main

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
	Tassert(interval.Start() == start, "start time: expected %v, got %v", start, interval.Start())
	Tassert(interval.End() == end, "end time: expected %v, got %v", end, interval.End())
}
