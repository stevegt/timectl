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
func NewTree() *Tree {
	return &Tree{}
}

// Insert inserts a new interval into the tree.
func (t *Tree) Insert(interval *Interval) {
	// Implementation needs to be reviewed as per test description
	if t.interval == nil {
		t.interval = interval
		t.maxEnd = interval.End()
	} else {
		if interval.End().After(t.maxEnd) {
			t.maxEnd = interval.End()
		}

		if interval.Start().Before(t.interval.Start()) {
			if t.left == nil {
				t.left = &Tree{interval: interval, maxEnd: interval.End()}
			} else {
				t.left.Insert(interval)
			}
		} else {
			if t.right == nil {
				t.right = &Tree{interval: interval, maxEnd: interval.End()}
			} else {
				t.right.Insert(interval)
			}
		}
	}
}

// Conflicts checks for conflicts with a given interval.
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
