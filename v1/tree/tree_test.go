package tree

import (
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// Tree is an interval tree that stores intervals and allows for fast
// lookup of intervals given an interval.  All inserted intervals are
// stored in leaf nodes; internal nodes store intervals that span the
// intervals of their children.  This allows for fast lookup of
// intervals that overlap a given interval.  Upon insertion of an
// interval, an existing leaf node is split into two leaf nodes, the
// existing interval is moved to one of the two leaf nodes, and the
// new interval is inserted into the other leaf node.  Left node
// intervals always have start times less than or equal to the right
// node interval start times.  If the start times are equal, the left
// node interval end time is less than or equal to the right node
// interval end time.

// test accumulator
func TestAccumulator(t *testing.T) {
	top := NewTree()

	// accumulate collects intervals in the tree that overlap the given
	// interval.  The intervals are collected in order of start time.

	Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:15:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T10:15:00Z")
	Ck(err)

	// get the intervals that overlap the range
	c1 := top.accumulate(searchStart, searchEnd)
	intervals := chan2slice(c1)

	// check that we got the right number of intervals
	Tassert(t, len(intervals) == 3, "Expected 3 intervals, got %d", len(intervals))

}

// test filter
func TestFilter(t *testing.T) {
	top := NewTree()

	// filter returns a channel of intervals from the input channel
	// that pass the filter function.

	Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)

	fn := func(interval interval.Interval) bool {
		return interval.Priority() < 2
	}

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:15:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T10:15:00Z")
	Ck(err)

	c1 := top.accumulate(searchStart, searchEnd)
	c2 := filter(c1, fn)
	i2 := chan2slice(c2)

	// check that we got the right number of intervals
	Tassert(t, len(i2) == 2, "Expected 2 intervals, got %d", len(i2))

}

// test contiguous filter
func TestContiguousFilter(t *testing.T) {
	top := NewTree()

	Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)
	Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 2)
	Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Insert(top, "2024-01-01T12:15:00Z", "2024-01-01T13:00:00Z", 1)

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:45:00Z")
	Ck(err)

	// get the intervals that overlap the range
	acc := top.accumulate(searchStart, searchEnd)
	// filter the intervals to only include those with a priority less than 2
	low := filter(acc, func(interval interval.Interval) bool {
		return interval.Priority() < 2
	})
	// filter the intervals to only include those that are contiguous
	// for at least N minutes
	cont := contiguous(low, 120*time.Minute)
	res := chan2slice(cont)

	// check that we got the right number of intervals
	Tassert(t, len(res) == 4, "Expected 4 intervals, got %d", len(res))

	// check that we got the right intervals
	err = Match(res[0], "2024-01-01T11:00:00Z", "2024-01-01T11:30:00Z", 0)
	Tassert(t, err == nil, err)
	err = Match(res[1], "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Tassert(t, err == nil, err)
	err = Match(res[2], "2024-01-01T12:00:00Z", "2024-01-01T12:15:00Z", 0)
	Tassert(t, err == nil, err)
	err = Match(res[3], "2024-01-01T12:15:00Z", "2024-01-01T13:00:00Z", 1)
	Tassert(t, err == nil, err)
}

// test rebalancing the tree
func TestRebalance(t *testing.T) {
	top := NewTree()

	// insert a few intervals into the tree
	Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)
	Insert(top, "2024-01-01T14:00:00Z", "2024-01-01T15:00:00Z", 1)

	// rebalance the tree
	top.rebalance()

	err := top.Verify()
	Tassert(t, err == nil, err)

	Verify(t, top, false)

}

// XXX WIP below here
