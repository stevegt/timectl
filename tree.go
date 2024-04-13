package interval

import "time"

// Tree represents a node in an interval tree.
type Tree struct {
	interval *Interval // The interval represented by this node
	left     *Tree     // Pointer to the left child
	right    *Tree     // Pointer to the right child
}

// NewTree creates and returns a new Tree node with a nil interval.
func NewTree() *Tree {
	return &Tree{}
}

// Insert adds a new interval to the tree, adjusting the structure as necessary.
func (t *Tree) Insert(newInterval *Interval) {
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
			t.left.Insert(newInterval)
		} else {
			if t.right == nil {
				t.right = &Tree{}
			}
			t.right.Insert(newInterval)
		}
	}
	t.updateSpanningInterval() // Update the interval to span its children.
}

// Conflicts finds and returns intervals in the tree that overlap with the given interval.
func (t *Tree) Conflicts(interval *Interval) []*Interval {
	var conflicts []*Interval
	if t.interval != nil && t.interval.Conflicts(interval) {
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
