package interval

import (
	"time"
	// . "github.com/stevegt/goadapt"
)

// TreeStart is the minimum time value that can be represented by a Tree node.
var TreeStart = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)

// TreeEnd is the maximum time value that can be represented by a Tree node.
var TreeEnd = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)

// TreeStartStr is the string representation of TreeStart.
var TreeStartStr = TreeStart.Format(time.RFC3339)

// TreeEndStr is the string representation of TreeEnd.
var TreeEndStr = TreeEnd.Format(time.RFC3339)

// Tree represents a node in an interval tree.
type Tree struct {
	// If this is not a leaf node, leafIinterval is nil.
	leafInterval *Interval
	// maxGap is the maximum gap between left end and right start times
	// for all children in the subtree rooted at this node
	// maxGap time.Duration
	left  *Tree // Pointer to the left child
	right *Tree // Pointer to the right child
}

// NewTree creates and returns a new Tree node containing a free interval spanning all time.
func NewTree() *Tree {
	return &Tree{leafInterval: NewInterval(TreeStart, TreeEnd, nil)}
}

// Insert adds a new interval to the tree, adjusting the structure as
// necessary.  Insertion fails if the new interval conflicts with any
// existing interval in the tree.  The new interval must be busy.
func (t *Tree) Insert(newInterval *Interval) bool {
	if !newInterval.Busy() {
		return false
	}

	if t.Busy() {
		if t.left != nil && newInterval.Start().Before(t.left.End()) {
			if t.left.Insert(newInterval) {
				return true
			}
		}
		if t.right != nil && newInterval.End().After(t.right.Start()) {
			if t.right.Insert(newInterval) {
				return true
			}
		}
		return false
	}

	// this is a free node, possibly with children -- we're going to
	// completely replace it with the results of punching a hole in
	// this node's interval
	newIntervals := t.Interval().Punch(newInterval)
	switch len(newIntervals) {
	case 0:
		// newInterval doesn't fit in this node's interval
		return false
	case 1:
		// newInterval fits exactly in this node's interval
		t.leafInterval = newInterval
		// clear out any children
		t.left = nil
		t.right = nil
		return true
	case 2:
		// newInterval fits in this node's interval, but we need to
		// split this node into two children
		t.leafInterval = nil
		t.left = &Tree{leafInterval: newIntervals[0]}
		t.right = &Tree{leafInterval: newIntervals[1]}
		return true
	case 3:
		// newInterval fits in this node's interval, but we need to
		// split this node into three children
		t.leafInterval = nil
		t.left = &Tree{leafInterval: newIntervals[0]}
		t.right = &Tree{
			left:  &Tree{leafInterval: newIntervals[1]},
			right: &Tree{leafInterval: newIntervals[2]},
		}
		return true
	default:
		panic("unexpected number of intervals")
	}
}

/*
// conflicts returns a channel containing intervals in leaf nodes that overlap with the given interval.
func (t *Tree) conflicts(left bool, interval *Interval) (out chan *Interval) {
	out = make(chan *Interval)
	go func(in *Interval) {
		defer close(out)
		// Pf("conflicts: in=%v, left=%v, right=%v\n", in, t.left, t.right)
		if t.leafInterval == nil {
			return
		}
		if t.left == nil && t.right == nil && t.leafInterval.Conflicts(in) {
			out <- t.leafInterval
			return
		}
		// Pf("conflicts: left=%v, right=%v\n", t.left, t.right)
		children := []*Tree{t.left, t.right}
		if !left {
			children = []*Tree{t.right, t.left}
		}
		for _, child := range children {
			if child == nil {
				continue
			}
			for conflict := range child.conflicts(left, in) {
				out <- conflict
			}
		}
	}(interval)
	return
}

// Conflicts returns a slice of intervals in leaf nodes that overlap with the given interval.
func (t *Tree) Conflicts(interval *Interval) []*Interval {
	var conflicts []*Interval
	for conflict := range t.conflicts(true, interval) {
		conflicts = append(conflicts, conflict)
	}
	return conflicts
}

// Conflict returns true if the given interval conflicts with any interval in the tree.
// If left is true, the left child is searched first; otherwise the right child is searched first.
func (t *Tree) Conflict(left bool, interval *Interval) bool {
	for range t.conflicts(left, interval) {
		return true
	}
	return false
}
*/

/*
// updateSpanningInterval updates the interval of this node to span its children.
func (t *Tree) updateSpanningInterval() {
	if t.left != nil && t.right != nil {
		busy := t.left.interval.Busy() || t.right.interval.Busy()
		t.interval = NewInterval(minTime(t.left.interval.Start(), t.right.interval.Start()), maxTime(t.left.interval.End(), t.right.interval.End()), busy)
	} else if t.left != nil {
		t.interval = t.left.interval
	} else if t.right != nil {
		t.interval = t.right.interval
	}
}
*/

/*
// updateMaxGap updates the maximum gap between left end and right start times.
func (t *Tree) updateMaxGap() {
	if t.left != nil && t.right != nil {
		gap := t.right.interval.Start().Sub(t.left.interval.End())
		t.maxGap = maxDuration(gap, maxDuration(t.left.maxGap, t.right.maxGap))
		// Pf("updateMaxGap: gap=%v, maxGap=%v\n", gap, t.maxGap)
		// Pf("updateMaxGap: left=%v, right=%v\n", t.left.interval, t.right.interval)
	}
}
*/

// minTime returns the earlier of two time.Time values.
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// maxTime returns the later of two time.Time values.
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// maxDuration returns the longer of two time.Duration values.
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// Intervals returns a slice of all intervals in all leaf nodes of the tree.
func (t *Tree) Intervals() []*Interval {
	var intervals []*Interval
	if t.left == nil && t.right == nil {
		intervals = append(intervals, t.leafInterval)
	}
	if t.left != nil {
		intervals = append(intervals, t.left.Intervals()...)
	}
	if t.right != nil {
		intervals = append(intervals, t.right.Intervals()...)
	}
	return intervals
}

// Busy returns true if any interval in the tree is busy.
func (t *Tree) Busy() bool {
	if t.leafInterval != nil && t.leafInterval.Busy() {
		return true
	}
	if t.left != nil && t.left.Busy() {
		return true
	}
	if t.right != nil && t.right.Busy() {
		return true
	}
	return false
}

// Start returns the start time of the interval spanning all child nodes.
func (t *Tree) Start() time.Time {
	if t.leafInterval != nil {
		return t.leafInterval.Start()
	}
	return t.left.Start()
}

// End returns the end time of the interval spanning all child nodes.
func (t *Tree) End() time.Time {
	if t.leafInterval != nil {
		return t.leafInterval.End()
	}
	return t.right.End()
}

// Interval returns the node's interval if interval is a leaf node, or
// returns a synthetic interval spanning all child nodes.
func (t *Tree) Interval() *Interval {
	if t.left == nil && t.right == nil {
		return t.leafInterval
	}
	return NewInterval(t.Start(), t.End(), t.Busy())
}
