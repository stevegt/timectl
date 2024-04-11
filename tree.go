package interval

import "time"

// Tree represents a node in an interval tree.
type Tree struct {
	interval *Interval // The interval represented by this node
	left     *Tree     // Pointer to the left child
	right    *Tree     // Pointer to the right child
}

// NewTree creates and returns a new Tree node without an interval.
func NewTree() *Tree {
	return &Tree{}
}

// Insert adds a new interval to the tree, adjusting the structure as necessary.
func (t *Tree) Insert(newInterval *Interval) {
	if t.interval == nil {
		t.interval = newInterval
		return
	}

	startNew := newInterval.Start()
	endNew := newInterval.End()

	if t.left == nil && t.right == nil {
		t.left = &Tree{interval: t.interval}
		if startNew.Before(t.interval.Start()) || (startNew.Equal(t.interval.Start()) && endNew.Before(t.interval.End())) {
			t.left.interval = newInterval
			t.interval = NewInterval(startNew, maxTime(t.interval.End(), endNew))
		} else {
			t.right = &Tree{interval: newInterval}
			t.interval = NewInterval(t.interval.Start(), maxTime(endNew, t.interval.End()))
		}
		return
	}

	if t.left != nil && startNew.Before(t.interval.Start()) {
		t.left.Insert(newInterval)
	} else if t.right != nil {
		t.right.Insert(newInterval)
	} else {
		t.right = &Tree{interval: newInterval}
	}

	t.interval = NewInterval(minTime(t.left.interval.Start(), t.right.interval.Start()), maxTime(t.left.interval.End(), t.right.interval.End()))
}

// Conflicts finds and returns intervals in the tree that overlap with the given interval.
func (t *Tree) Conflicts(interval *Interval) []*Interval {
	var conflicts []*Interval
	if t.interval == nil {
		return conflicts
	}

	if t.interval.Conflicts(interval) {
		if t.left == nil && t.right == nil {
			conflicts = append(conflicts, t.interval)
		}
		if t.left != nil {
			conflicts = append(conflicts, t.left.Conflicts(interval)...)
		}
		if t.right != nil {
			conflicts = append(conflicts, t.right.Conflicts(interval)...)
		}
	}

	return conflicts
}

// minTime returns the minimum between two time.Time values
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// maxTime returns the maximum between two time.Time values
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
