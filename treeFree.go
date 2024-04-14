package interval

import (
	"time"

	. "github.com/stevegt/goadapt"
)

// FirstFree walks the tree to find the first free interval that
// is at least the given duration.  The searchStart and searchEnd
// times are inclusive.  The duration is exclusive.  The search
// uses an internal walk() function that recursively walks the
// tree in a depth-first manner, following the left child first.
// It returns the first free interval that it finds, or nil if
// no free interval is found.
func (t *Tree) FirstFree(minStart, maxEnd time.Time, duration time.Duration) *Interval {
	return findFirst(t, minStart, maxEnd, duration)
}

func findFirst(node *Tree, start, end time.Time, duration time.Duration) *Interval {
	if start.Add(duration).After(end) {
		// if the duration is longer than the time between start and end,
		// then we can't find a free interval
		return nil
	}

	// If the given node or its interval is nil, then there are no
	// intervals in this subtree, so we can create a free interval
	// at the start time
	if node == nil || node.interval == nil {
		return NewInterval(start, start.Add(duration))
	}

	busy := node.interval
	busyStart := busy.Start()
	busyEnd := busy.End()

	// if the children are nil, then the current node is a leaf node
	isLeaf := node.left == nil && node.right == nil
	// if the maxGap is too small, then we won't find a free interval
	// in this subtree
	isFull := node.maxGap < duration
	if isLeaf || isFull {
		// try fitting a free interval before the busy interval
		before := findFirst(nil, start, busyStart, duration)
		if before != nil {
			return before
		}
		// try fitting a free interval after the busy interval
		after := findFirst(nil, busyEnd, end, duration)
		if after != nil {
			return after
		}
		return nil
	}

	// everything after here is for non-leaf nodes
	Assert(node.left != nil, "left child is nil")
	Assert(node.right != nil, "right child is nil")
	rightStart := node.right.interval.Start()

	// drill down the left subtree
	leftResult := findFirst(node.left, start, rightStart, duration)
	if leftResult != nil {
		return leftResult
	}

	// drill down the right subtree
	return findFirst(node.right, rightStart, end, duration)
}

// find returns an interval that has the given duration.  The interval
// starts as early as possible if first is true, and as late as possible
// if first is false.  The minStart and maxEnd times are inclusive.
// The duration is exclusive.
//
// This function works by walking the tree in a depth-first manner,
// following the left child first if first is set, otherwise following
// the right child first.
