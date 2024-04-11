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
	// Setting up the initial node if the tree is empty
	if t.interval == nil {
		t.interval = interval
		t.maxEnd = interval.End()
		return
	}

	// Update maxEnd if the current interval ends later
	if interval.End().After(t.maxEnd) {
		t.maxEnd = interval.End()
	}

	// Insert based purely on start time comparison might be oversimplifying,
	// inserting to the left if it starts before the current interval
	if interval.Start().Before(t.interval.Start()) {
		if t.left == nil {
			t.left = &Tree{}
		}
		t.left.Insert(interval)
	} else {
		// And to the right if it starts at the same time or after
		if t.right == nil {
			t.right = &Tree{}
		}
		t.right.Insert(interval)
	}
}

// Conflicts checks for conflicts with a given interval.
func (t *Tree) Conflicts(interval *Interval) []*Interval {
	var conflicts []*Interval
	if t.interval == nil {
		return conflicts
	}

	// Check the current node for a conflict
	if t.interval.Conflicts(interval) {
		conflicts = append(conflicts, t.interval)
	}

	// Traverse the left subtree if it exists and might intersect
	if t.left != nil && t.left.maxEnd.After(interval.Start()) {
		conflicts = append(conflicts, t.left.Conflicts(interval)...)
	}

	// And the right subtree
	if t.right != nil && interval.End().After(t.interval.Start()) {
		conflicts = append(conflicts, t.right.Conflicts(interval)...)
	}

	return conflicts
}
