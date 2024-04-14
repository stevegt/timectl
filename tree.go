package interval

import (
	"time"
	// . "github.com/stevegt/goadapt"
)

// Tree represents a node in an interval tree.
type Tree struct {
	// If this is not a leaf node, interval is the interval that spans
	// all child nodes -- i.e. the minimum start time and maximum end time
	// of all intervals in the subtree rooted at this node.
	interval *Interval
	// maxGap is the maximum gap between left end and right start times
	// for all children in the subtree rooted at this node
	maxGap time.Duration
	left   *Tree // Pointer to the left child
	right  *Tree // Pointer to the right child
}

// NewTree creates and returns a new Tree node with a nil interval.
func NewTree() *Tree {
	return &Tree{}
}

// Insert adds a new interval to the tree, adjusting the structure as
// necessary.  Insertion fails if the new interval conflicts with any
// existing interval in the tree.
func (t *Tree) Insert(newInterval *Interval) bool {
	if t.Conflict(true, newInterval) {
		return false
	}
	t.Insert(newInterval)
	return true
}

func (t *Tree) insert(newInterval *Interval) {

	if t.interval == nil && t.left == nil && t.right == nil {
		t.interval = newInterval
		return
	}

	if t.left == nil && t.right == nil {
		if newInterval.Start().Before(t.interval.Start()) {
			t.left = &Tree{interval: newInterval}
			t.right = &Tree{interval: t.interval}
		} else {
			t.left = &Tree{interval: t.interval}
			t.right = &Tree{interval: newInterval}
		}
		t.interval = nil // Clear current node's interval as it's now an internal node.
	} else {
		// Determine whether to insert into left or right subtree.
		if newInterval.Start().Before(t.interval.Start()) || (newInterval.Start().Equal(t.interval.Start()) && newInterval.End().Before(t.interval.End())) {
			if t.left == nil {
				t.left = &Tree{}
			}
			t.left.insert(newInterval)
		} else {
			if t.right == nil {
				t.right = &Tree{}
			}
			t.right.insert(newInterval)
		}
	}
	t.updateSpanningInterval() // Update the interval to span its children.
	t.updateMaxGap()           // Update the maximum gap between left end and right start times.
}

// conflicts returns a channel containing intervals in leaf nodes that overlap with the given interval.
func (t *Tree) conflicts(left bool, interval *Interval) (out chan *Interval) {
	out = make(chan *Interval)
	go func() {
		defer close(out)
		if t.left == nil && t.right == nil && t.interval.Conflicts(interval) {
			out <- t.interval
		}
		children := []*Tree{t.left, t.right}
		if !left {
			children = []*Tree{t.right, t.left}
		}
		for _, child := range children {
			if child != nil {
				for conflict := range child.conflicts(left, interval) {
					out <- conflict
				}
			}
		}
	}()
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

// updateSpanningInterval updates the interval of this node to span its children.
func (t *Tree) updateSpanningInterval() {
	if t.left != nil && t.right != nil {
		t.interval = NewInterval(minTime(t.left.interval.Start(), t.right.interval.Start()), maxTime(t.left.interval.End(), t.right.interval.End()))
	} else if t.left != nil {
		t.interval = t.left.interval
	} else if t.right != nil {
		t.interval = t.right.interval
	}
}

// updateMaxGap updates the maximum gap between left end and right start times.
func (t *Tree) updateMaxGap() {
	if t.left != nil && t.right != nil {
		gap := t.right.interval.Start().Sub(t.left.interval.End())
		t.maxGap = maxDuration(gap, maxDuration(t.left.maxGap, t.right.maxGap))
		// Pf("updateMaxGap: gap=%v, maxGap=%v\n", gap, t.maxGap)
		// Pf("updateMaxGap: left=%v, right=%v\n", t.left.interval, t.right.interval)
	}
}

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
		intervals = append(intervals, t.interval)
	}
	if t.left != nil {
		intervals = append(intervals, t.left.Intervals()...)
	}
	if t.right != nil {
		intervals = append(intervals, t.right.Intervals()...)
	}
	return intervals
}
