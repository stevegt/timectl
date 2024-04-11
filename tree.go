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
	if t.interval == nil && t.left == nil && t.right == nil {
		t.interval = newInterval
		return
	}

	// Decide whether to insert into left or right subtree based on the start time.
	start := newInterval.Start()
	if start.Before(t.interval.Start()) || start.Equal(t.interval.Start()) && newInterval.End().Before(t.interval.End()) {
		if t.left == nil {
			t.left = NewTree()
		}
		t.left.Insert(newInterval)
	} else {
		if t.right == nil {
			t.right = NewTree()
		}
		t.right.Insert(newInterval)
	}
	t.updateSpanningInterval() // Make sure to update the spanning interval to include new child nodes.
}

// Conflicts finds and returns intervals in the tree that overlap with the given interval.
func (t *Tree) Conflicts(interval *Interval) []*Interval {
	var conflicts []*Interval
	if t.interval == nil {
		return conflicts
	}

	if t.interval.Conflicts(interval) {
		conflicts = append(conflicts, t.interval)
	}

	if t.left != nil {
		conflicts = append(conflicts, t.left.Conflicts(interval)...)
	}
	if t.right != nil {
		conflicts = append(conflicts, t.right.Conflicts(interval)...)
	}

	return conflicts
}

// updateSpanningInterval updates the interval for the node to span its children,
// creating a new interval that covers both child intervals when necessary.
func (t *Tree) updateSpanningInterval() {
	// Only need to update if the node is not a leaf node.
	if t.left != nil || t.right != nil {
		var start, end time.Time

		if t.left != nil && t.right != nil {
			start = minTime(t.left.interval.Start(), t.right.interval.Start())
			end = maxTime(t.left.interval.End(), t.right.interval.End())
		} else if t.left != nil {
			start = t.left.interval.Start()
			end = t.left.interval.End()
		} else if t.right != nil {
			start = t.right.interval.Start()
			end = t.right.interval.End()
		}

		// Update the node's interval.
		t.interval = NewInterval(start, end)
	}
}

// minTime returns the earlier of two time.Time values.
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// maxTime returns the latter of two time.Time values.
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
