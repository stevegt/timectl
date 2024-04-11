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

// NewTree creates and returns a new Tree node without an interval.
// This change will help in ensuring that our tree starts off empty,
// and the root node will be populated on the first insertion.
func NewTree() *Tree {
	return &Tree{}
}

// Insert inserts a new interval into the tree.
func (t *Tree) Insert(interval *Interval) {
	// If the node is empty (has no interval), populate it with the new interval and set the maxEnd.
	if t.interval == nil {
		t.interval = interval
		t.maxEnd = interval.End()
		return
	}

	// Update the maxEnd at each node to ensure it represents the maximum end time in its subtree.
	if interval.End().After(t.maxEnd) {
		t.maxEnd = interval.End()
	}

	// Determine if the new interval goes to the left or right,
	// and perform insertion recursively.
	if interval.Start().Before(t.interval.Start()) {
		// If the left child does not exist, create it and insert the interval there.
		if t.left == nil {
			t.left = &Tree{interval: interval, maxEnd: interval.End()}
		} else {
			// Else, recursively insert into the left subtree.
			t.left.Insert(interval)
		}
	} else {
		// Similar to the left insertion but for the right child.
		if t.right == nil {
			t.right = &Tree{interval: interval, maxEnd: interval.End()}
		} else {
			// Else, recursively insert into the right subtree.
			t.right.Insert(interval)
		}
	}
}

// Conflicts checks for conflicts with a given interval within the tree.
func (t *Tree) Conflicts(interval *Interval) []*Interval {
	// Check if the current node is empty.
	var conflicts []*Interval
	if t.interval == nil {
		return conflicts
	}

	// If the current interval conflicts with the given one, add it to the result.
	if t.interval.Conflicts(interval) {
		conflicts = append(conflicts, t.interval)
	}

	// If there's a left child and there could be a conflict, check the left subtree.
	if t.left != nil && t.left.maxEnd.After(interval.Start()) {
		conflicts = append(conflicts, t.left.Conflicts(interval)...)
	}

	// Likewise for the right subtree.
	if t.right != nil && interval.End().After(t.interval.Start()) {
		conflicts = append(conflicts, t.right.Conflicts(interval)...)
	}

	return conflicts
}
