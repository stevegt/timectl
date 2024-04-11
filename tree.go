package interval

import (
	"time"
)

// Tree represents a node in an interval tree.
type Tree struct {
	interval *Interval // The interval stored in this node
	maxEnd   time.Time // The maximum end time in this subtree
	left     *Tree     // Pointer to the left child
	right    *Tree     // Pointer to the right child
}

// NewTree creates and returns a new Tree.
// Initially, it doesn't contain any intervals, hence no need for parameters.
func NewTree() *Tree {
	return &Tree{}
}

// Insert inserts a new interval into the tree.
func (t *Tree) Insert(interval *Interval) {
	if t.interval == nil {
		t.interval = interval
		t.maxEnd = interval.End()
		return
	}

	if interval.End().After(t.maxEnd) {
		t.maxEnd = interval.End()
	}

	if interval.Start().Before(t.interval.Start()) {
		if t.left == nil {
			t.left = NewTree()
		}
		t.left.Insert(interval)
	} else {
		if t.right == nil {
			t.right = NewTree()
		}
		t.right.Insert(interval)
	}
}

// Conflicts returns a slice of intervals that conflict with the given interval.
func (t *Tree) Conflicts(interval *Interval) []*Interval {
	var conflicts []*Interval
	if t.interval == nil {
		return conflicts
	}

	if t.interval.Conflicts(interval) {
		conflicts = append(conflicts, t.interval)
	}

	if t.left != nil && t.left.maxEnd.After(interval.Start()) {
		conflicts = append(conflicts, t.left.Conflicts(interval)...)
	}

	if t.right != nil && interval.End().After(t.interval.Start()) {
		conflicts = append(conflicts, t.right.Conflicts(interval)...)
	}

	return conflicts
}
