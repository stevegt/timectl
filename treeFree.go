package interval

import (
	"fmt"
	"time"
	// . "github.com/stevegt/goadapt"
)

/*
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
*/

/*
// freeSlots returns at most three intervals:
//  1. A free interval that starts at the minStart time and ends at the
//     start of the busy interval in the node.
//  2. A free interval that starts at the end of the left child's busy
//     interval and ends at the start of the right child's busy interval.
//  3. A free interval that starts at the end of the busy interval in the
//     node and ends at the maxEnd time.
func (node *Tree) freeSlots(minStart, maxEnd time.Time) (intervals []*Interval) {
	if node == nil || node.interval == nil {
		intervals = append(intervals, NewInterval(minStart, maxEnd))
		return
	}

	busy := node.interval
	busyStart := busy.Start()
	busyEnd := busy.End()

	Pf("busyStart: %v, minStart: %v busyEnd: %v, maxEnd: %v\n", busyStart, minStart, busyEnd, maxEnd)

	if busyStart.After(minStart) {
		intervals = append(intervals, NewInterval(minStart, busyStart))
	}

	if node.left != nil && node.right != nil {
		leftEnd := node.left.interval.End()
		rightStart := node.right.interval.Start()
		Pf("leftEnd: %v, rightStart: %v\n", leftEnd, rightStart)
		intervals = append(intervals, NewInterval(leftEnd, rightStart))
	}

	if busyEnd.Before(maxEnd) {
		intervals = append(intervals, NewInterval(busyEnd, maxEnd))
	}

	return
}
*/

/*
// genslots returns a channel of free intervals that are generated
// by walking the tree in a depth-first manner.  The minStart and
// maxEnd times are inclusive.  The duration is exclusive. If first
// is true, then the intervals are generated in order from the earliest
// start time to the latest start time.  If first is false, then the
// intervals are generated in order from the latest start time to the
// earliest start time.
func (t *Tree) genSlots(first bool, minStart, maxEnd time.Time) (out chan *Interval) {
	out = make(chan *Interval)
	go func() {
		defer close(out)
		if maxEnd.Sub(minStart) <= 0 {
			return
		}

		// if the given node or its interval is nil, then there are no
		// intervals in this subtree, so we can create a free interval
		// at the start time
		if t == nil || t.leafInterval == nil {
			out <- NewInterval(minStart, maxEnd, nil)
			return
		}

		// figure out the sequence to walk the tree
		seq := []string{"pre", "left", "gap", "right", "post"}
		if !first {
			seq = []string{"post", "right", "gap", "left", "pre"}
		}

		// walk the tree in a depth-first manner according to seq
		for _, dir := range seq {
			var slot *Interval
			switch dir {
			case "pre":
				start := minStart
				end := minTime(t.leafInterval.Start(), maxEnd)
				slot = NewInterval(start, end, nil)
				out <- slot
			case "left":
				if t.left != nil {
					start := maxTime(minStart, t.left.leafInterval.Start())
					end := minTime(t.left.leafInterval.End(), maxEnd)
					for slot := range t.left.genSlots(first, start, end) {
						if slot == nil {
							continue
						}
						if t.Conflict(first, slot) {
							Pf("conflict: %v\n", slot)
							dump(t, 0)
							Assert(false, "conflict")
						}
						out <- slot
					}
				}
			case "gap":
				if t.left != nil && t.right != nil {
					start := maxTime(minStart, t.left.leafInterval.End())
					end := minTime(t.right.leafInterval.Start(), maxEnd)
					slot = NewInterval(start, end, nil)
					out <- slot
				}
			case "right":
				if t.right != nil {
					start := maxTime(minStart, t.right.leafInterval.Start())
					end := minTime(t.right.leafInterval.End(), maxEnd)
					for slot := range t.right.genSlots(first, start, end) {
						if slot == nil {
							continue
						}
						if t.Conflict(first, slot) {
							Pf("slot: %v\n", slot)
							for conflict := range t.conflicts(first, slot) {
								Pf("conflict: %v\n", conflict)
							}
							dump(t, 0)
							Assert(false, "conflict")
						}
						out <- slot
					}
				}
			case "post":
				start := maxTime(minStart, t.leafInterval.End())
				end := maxEnd
				slot = NewInterval(start, end, nil)
				out <- slot
			}
		}
	}()
	return
}
*/

// FindFree returns a free interval that has the given duration.  The
// interval starts as early as possible if first is true, and as late
// as possible if first is false.  The minStart and maxEnd times are
// inclusive. The duration is exclusive.
//
// This function works by walking the tree in a depth-first manner,
// following the left child first if first is set, otherwise following
// the right child first.
func (t *Tree) FindFree(first bool, minStart, maxEnd time.Time, duration time.Duration) (free *Interval) {

	if !t.Busy() {
		start := maxTime(minStart, t.Start())
		end := minTime(t.End(), maxEnd)
		return subInterval(first, start, end, duration)
	}

	var children []*Tree
	var start, end time.Time
	if first {
		children = []*Tree{t.left, t.right}
	} else {
		children = []*Tree{t.right, t.left}
	}

	for _, child := range children {
		if child == nil {
			continue
		}
		start = maxTime(minStart, child.Start())
		end = minTime(child.End(), maxEnd)
		slot := child.FindFree(first, start, end, duration)
		if slot != nil {
			return slot
		}
	}

	// if we get here, then we didn't find a free interval anywhere
	// under this node
	return nil
}

// subInterval returns a free interval that starts as early as possible
// if first is true, and as late as possible if first is false.  The
// minStart and maxEnd times are inclusive. The duration is exclusive.
// If the duration is longer than the time between minStart and maxEnd,
// then this function returns nil.
func subInterval(first bool, minStart, maxEnd time.Time, duration time.Duration) *Interval {
	if maxEnd.Sub(minStart) < duration {
		return nil
	}
	if first {
		return NewInterval(minStart, minStart.Add(duration), nil)
	}
	return NewInterval(maxEnd.Add(-duration), maxEnd, nil)
}

// dump is a helper function that prints the tree structure to
// stdout.
func dump(tree *Tree, path string) {
	// fmt.Printf("maxGap: %v interval: %v\n", tree.maxGap, tree.interval)
	fmt.Printf("%v: %v\n", path, tree.leafInterval)
	if tree.left != nil {
		dump(tree.left, path+"l")
	}
	if tree.right != nil {
		dump(tree.right, path+"r")
	}
}
