package tree

import (
	"fmt"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
	"github.com/stevegt/timectl/util"
)

// TreeStart is the minimum time value that can be represented by a Tree node.
var TreeStart = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)

// TreeEnd is the maximum time value that can be represented by a Tree node.
var TreeEnd = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)

// TreeStartStr is the string representation of TreeStart.
var TreeStartStr = TreeStart.Format(time.RFC3339)

// TreeEndStr is the string representation of TreeEnd.
var TreeEndStr = TreeEnd.Format(time.RFC3339)

// NewTree creates and returns a new Tree node containing a free interval spanning all time.
func NewTree() *Node {
	return newNodeFromInterval(interval.NewInterval(TreeStart, TreeEnd, 0))
}

// SetLeft sets the left child of this node.  It returns the old left
// child or nil if there was no old left child.  If the given child node
// is already a child of another node, the right child of this node,
// or the parent of this node, then this function clears the old
// relationship before setting the new one.
func (t *Node) SetLeft(left *Node) (old *Node) {
	old = t.left
	if left != nil && left.parent != nil {
		if left.parent.left == left {
			left.parent.left = nil
		}
		if left.parent.right == left {
			left.parent.right = nil
		}
	}
	if t.right == left {
		t.right = nil
	}
	t.left = left
	if t.left != nil {
		t.left.parent = t
		t.left.setMinMax()
	} else {
		t.setMinMax()
	}
	return
}

// SetRight sets the right child of this node.  It returns the old right
// child or nil if there was no old right child.  If the given child node
// is already a child of another node, the left child of this node,
// or the parent of this node, then this function clears the old
// relationship before setting the new one.
func (t *Node) SetRight(right *Node) (old *Node) {
	old = t.right
	if right != nil && right.parent != nil {
		if right.parent.left == right {
			right.parent.left = nil
		}
		if right.parent.right == right {
			right.parent.right = nil
		}
	}
	if t.left == right {
		t.left = nil
	}
	t.right = right
	if t.right != nil {
		t.right.parent = t
		t.right.setMinMax()
	} else {
		t.setMinMax()
	}
	return
}

// Insert adds a new interval to the tree, adjusting the structure as
// necessary.  Insertion fails if the new interval conflicts with any
// existing interval in the tree with a priority greater than 0.
// Insertion fails if the new interval is not busy.
func (t *Node) Insert(newInterval interval.Interval) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !newInterval.Busy() {
		// XXX return a meaningful error
		return false
	}

	// use FindLowerPriority to find a free interval where we can
	// insert the new interval
	nodes := t.FindLowerPriority(true, newInterval.Start(), newInterval.End(), newInterval.Duration(), 1)
	if len(nodes) == 0 {
		// XXX return a meaningful error
		return false
	}

	// nodes should be a slice of length 1
	Assert(len(nodes) == 1, "unexpected number of nodes")
	f := nodes[0]
	// freeNode should have a free interval
	Assert(!f.Busy(), "freeNode is busy")

	// f should start on or before newInterval and end on or after
	// newInterval
	if f.Start().After(newInterval.Start()) || f.End().Before(newInterval.End()) {
		// XXX return a meaningful error
		return false
	}

	newIntervals := f.Interval().Punch(newInterval)
	switch len(newIntervals) {
	case 0:
		// newInterval doesn't fit in this node's interval
		Assert(false, "newInterval doesn't fit in freeNode's interval")
	case 1:
		// newInterval fits exactly in this node's interval
		f.SetInterval(newInterval)
		f.setMinMax()
		return true
	case 2:
		// newInterval fits in this node's interval with a free interval
		// left over
		// put the first interval in this node
		f.SetInterval(newIntervals[0])
		// create a new right child for the second interval and make
		// the old right child the right child of it
		newNode := newNodeFromInterval(newIntervals[1])
		oldRight := f.SetRight(newNode)
		f.right.SetRight(oldRight)
		return true
	case 3:
		// newInterval fits in this node's interval with free intervals
		// remaining to the left and right, so...

		// put the first interval in a new left child, moving the old
		// left child to the left of the new left child
		newLeftNode := newNodeFromInterval(newIntervals[0])
		oldLeft := f.SetLeft(newLeftNode)
		f.left.SetLeft(oldLeft)

		// put the second interval in this node
		f.SetInterval(newIntervals[1])

		// put the third interval in a new right child, moving the old
		// right child to the right of the new right child
		newRightNode := newNodeFromInterval(newIntervals[2])
		oldRight := f.SetRight(newRightNode)
		f.right.SetRight(oldRight)
		return true
	default:
		Assert(false, "unexpected number of intervals")
	}

	// check everything
	// XXX either remove this or refactor all of the above to use
	// Height in the first place and not need rebalancing
	f.ckHeight()
	f.right.ckHeight()
	f.left.ckHeight()

	Assert(false, "unexpected code path")
	return false
}

// ckHeight checks the calculated height of the node against the actual
// height of the node's subtree.
func (t *Node) ckHeight() {
	if t == nil {
		return
	}
	calculatedHeight := t.Height()
	actualHeight := t.CalcHeight()
	Assert(calculatedHeight == actualHeight, "height mismatch Height: %d height(): %d", calculatedHeight, actualHeight)
}

// BusyIntervals returns a slice of all busy intervals in all leaf nodes of the tree.
func (t *Node) BusyIntervals() (intervals []interval.Interval) {
	t.mu.Lock()
	defer t.mu.Unlock()
	// XXX inefficient -- use MaxPriority
	for _, i := range t.AllIntervals() {
		if i.Busy() {
			intervals = append(intervals, i)
		}
	}
	return
}

// AllIntervals returns a slice of all intervals in all leaf nodes of the tree.
func (t *Node) AllIntervals() []interval.Interval {
	t.mu.Lock()
	defer t.mu.Unlock()

	var intervals []interval.Interval
	if t.left != nil {
		intervals = append(intervals, t.left.AllIntervals()...)
	}
	intervals = append(intervals, t.Interval())
	if t.right != nil {
		intervals = append(intervals, t.right.AllIntervals()...)
	}
	return intervals
}

// Conflicts returns a slice of intervals in leaf nodes that overlap with the given interval.
// If includeFree is true, then this function returns all intervals that conflict with the given
// interval, otherwise it returns only busy intervals.
func (t *Node) Conflicts(iv interval.Interval, includeFree bool) []interval.Interval {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t == nil {
		return nil
	}

	var conflicts []interval.Interval
	if t.Interval().Conflicts(iv, includeFree) {
		conflicts = append(conflicts, t.Interval())
	}
	if t.left != nil {
		conflicts = append(conflicts, t.left.Conflicts(iv, includeFree)...)
	}
	if t.right != nil {
		conflicts = append(conflicts, t.right.Conflicts(iv, includeFree)...)
	}
	return conflicts
}

// FindFree returns a free interval that has the given duration.  The
// interval starts as early as possible if first is true, and as late
// as possible if first is false.  The minStart and maxEnd times are
// inclusive. The duration is exclusive.
//
// This function works by walking the tree in a depth-first manner,
// following the left child first if first is set, otherwise following
// the right child first.
func (t *Node) FindFree(first bool, minStart, maxEnd time.Time, duration time.Duration) (free interval.Interval) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Pf("FindFree: first: %v minStart: %v maxEnd: %v duration: %v\n", first, minStart, maxEnd, duration)
	// Pf("busy: %v\n", t.Busy())
	if !t.Busy() {
		start := util.MaxTime(minStart, t.minStart)
		end := util.MinTime(t.MaxEnd(), maxEnd)
		sub := subInterval(first, start, end, duration)
		return sub
	}

	var children []*Node
	var start, end time.Time
	if first {
		children = []*Node{t.left, t.right}
	} else {
		children = []*Node{t.right, t.left}
	}

	for _, child := range children {
		if child == nil {
			continue
		}
		start = util.MaxTime(minStart, child.minStart)
		end = util.MinTime(child.MaxEnd(), maxEnd)
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
func subInterval(first bool, minStart, maxEnd time.Time, duration time.Duration) interval.Interval {
	if maxEnd.Sub(minStart) < duration {
		return nil
	}
	if first {
		return interval.NewInterval(minStart, minStart.Add(duration), 0)
	}
	return interval.NewInterval(maxEnd.Add(-duration), maxEnd, 0)
}

// FreeIntervals returns a slice of all free intervals in all leaf nodes of the tree.
func (t *Node) FreeIntervals() (intervals []interval.Interval) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, i := range t.AllIntervals() {
		if !i.Busy() {
			intervals = append(intervals, i)
		}
	}
	return
}

// allPaths returns a channel of all paths to all nodes in the tree.
// The paths are sorted in depth-first order, left child first.
func (t *Node) allPaths(path Path) (c chan Path) {
	c = make(chan Path)
	go func() {
		defer close(c)
		t.allPathsBlocking(path, c)
	}()
	return c
}

// allPathsBlocking is a helper function for allPaths that returns a
// channel of all paths to all nodes in the tree.  The paths are sorted
// in depth-first order, left child first.
func (t *Node) allPathsBlocking(path Path, c chan Path) {
	myPath := path.Append(t)
	// Pf("path %p myPath %p\n", path, myPath)
	// Pf("send: %-10s %v\n", myPath, t.leafInterval)
	if t.left != nil {
		t.left.allPathsBlocking(myPath, c)
	}
	c <- myPath
	if t.right != nil {
		t.right.allPathsBlocking(myPath, c)
	}
}

// allNodes returns a channel of all nodes in the tree that are
// between the start and end times.  The fwd parameter determines
// whether the nodes are returned in depth-first order, left child
// first, or in reverse depth-first order, right child first.
func (t *Node) allNodes(fwd bool, start, end time.Time) <-chan *Node {
	c := make(chan *Node)
	go func() {
		defer close(c)
		t.allNodesBlocking(fwd, start, end, c)
	}()
	return c
}

// allNodesBlocking is a helper function for allNodes that returns a
// channel of all nodes in the tree.  The fwd parameter determines
// whether the nodes are returned in depth-first order, left child
// first, or in reverse depth-first order, right child first.
func (t *Node) allNodesBlocking(fwd bool, start, end time.Time, c chan *Node) {

	if t == nil {
		return
	}

	if t.MaxEnd().Before(start) {
		return
	}
	if t.minStart.After(end) {
		return
	}

	if fwd {
		t.left.allNodesBlocking(fwd, start, end, c)
		c <- t
		t.right.allNodesBlocking(fwd, start, end, c)
	} else {
		t.right.allNodesBlocking(fwd, start, end, c)
		c <- t
		t.left.allNodesBlocking(fwd, start, end, c)
	}
}

// FirstNode returns the first node in the tree.
func (t *Node) FirstNode() *Node {
	if t.left != nil {
		return t.left.FirstNode()
	}
	return t
}

// LastNode returns the last node in the tree.
func (t *Node) LastNode() *Node {
	if t.right != nil {
		return t.right.LastNode()
	}
	return t
}

// AsDot returns a string representation of the tree in Graphviz DOT
// format without relying on any other Tree methods.
func (t *Node) AsDot(path Path) string {
	// t.Mu.Lock()
	// defer t.Mu.Unlock()

	var out string
	var top bool
	if path == nil {
		top = true
		path = Path{t}
		out += "digraph G {\n"
	}
	id := path.String()
	label := Spf("parent %p\\nthis %p\\n", t.parent, t)
	label += Spf("left %p    right %p\\n", t.left, t.right)
	label += Spf("%v\\nminStart %v\\nmaxEnd %v\\nmaxPriority %v", id, t.minStart, t.MaxEnd(), t.MaxPriority())
	if t.Interval() != nil {
		label += fmt.Sprintf("\\n%s", t.Interval())
	} else {
		label += "\\nnil"
	}
	out += fmt.Sprintf("  %s [label=\"%s\"];\n", id, label)
	if t.left != nil {
		// get left child's dot representation
		out += t.left.AsDot(path.Append(t.left))
		// add edge from this node to left child
		out += fmt.Sprintf("  %s -> %sl [label=%s];\n", id, id, "l")
	} else {
		out += fmt.Sprintf("  %sl [label=\"nil\" style=dotted];\n", id)
		out += fmt.Sprintf("  %s -> %sl [label=%s];\n", id, id, "l")
	}
	if t.right != nil {
		// get right child's dot representation
		out += t.right.AsDot(path.Append(t.right))
		// add edge from this node to right child
		out += fmt.Sprintf("  %s -> %sr [label=%s];\n", id, id, "r")
	} else {
		out += fmt.Sprintf("  %sr [label=\"nil\" style=dotted];\n", id)
		out += fmt.Sprintf("  %s -> %sr [label=%s];\n", id, id, "r")
	}
	if top {
		out += "}\n"
	}
	return out
}

// rotateLeft performs a left rotation on this node.
func (t *Node) rotateLeft() (R *Node) {
	if t == nil || t.right == nil {
		return
	}
	// we start like this:
	//
	//       t
	//        \
	//         R
	//        / \
	//       x   y
	//
	R = t.right
	x := R.left

	// pivot around R
	//
	//         R
	//        / \
	//       t   y
	//		  \
	//		   x
	//
	R.left = t
	t.right = x
	R.parent = t.parent
	t.parent = R
	if R.parent != nil {
		switch {
		case R.parent.left == t:
			R.parent.left = R
		case R.parent.right == t:
			R.parent.right = R
		default:
			Assert(false, "can't find t in R.Parent")
		}
	}
	if x != nil {
		x.parent = t
		x.setMinMax()
	} else {
		t.setMinMax()
	}
	return
}

// rotateRight performs a right rotation on this node.
func (t *Node) rotateRight() (L *Node) {
	if t == nil || t.left == nil {
		return
	}
	// we start like this:
	//
	//       t
	//      /
	//     L
	//    / \
	//   x   y
	//
	L = t.left
	y := L.right

	// pivot around L
	//
	//     L
	//    / \
	//   x   t
	//      /
	//     y
	//
	L.right = t
	t.left = y
	L.parent = t.parent
	t.parent = L
	if L.parent != nil {
		switch {
		case L.parent.left == t:
			L.parent.left = L
		case L.parent.right == t:
			L.parent.right = L
		default:
			Assert(false, "can't find t in L.Parent")
		}
	}
	if y != nil {
		y.parent = t
		y.setMinMax()
	} else {
		t.setMinMax()
	}
	return
}

// setMinMax updates the minimum and maximum values of this node and
// its ancestors.
func (t *Node) setMinMax() {
	if t == nil {
		return
	}

	/*
		for _, s := range seen {
			if s == t {
				Pf("seen:\n")
				for _, s := range seen {
					Pf("%s\n", s)
				}
				Pf("t: %s\n", t)
				Assert(false, "cycle detected")
			}
		}

		seen = append(seen, t)
	*/

	var leftHeight, rightHeight int
	var leftSize, rightSize int
	if t.left == nil {
		t.SetMinStart(t.Interval().Start())
	} else {
		t.SetMinStart(t.left.minStart)
		leftHeight = t.left.Height()
		leftSize = t.left.Size()
	}
	if t.right == nil {
		t.SetMaxEnd(t.Interval().End())
	} else {
		t.SetMaxEnd(t.right.MaxEnd())
		rightHeight = t.right.Height()
		rightSize = t.right.Size()
	}

	t.SetMaxPriority(t.Interval().Priority())
	t.SetMinPriority(t.Interval().Priority())
	if t.left != nil {
		t.SetMaxPriority(max(t.MaxPriority(), t.left.MaxPriority()))
		t.SetMinPriority(min(t.MinPriority(), t.left.MinPriority()))
	}
	if t.right != nil {
		t.SetMaxPriority(max(t.MaxPriority(), t.right.MaxPriority()))
		t.SetMinPriority(min(t.MinPriority(), t.right.MinPriority()))
	}

	// the height of the node is the height of the tallest child plus 1
	t.SetHeight(1 + max(leftHeight, rightHeight))
	// the size of the node is the size of the left child plus the size
	// of the right child plus 1
	t.SetSize(1 + leftSize + rightSize)

	if t.parent != nil {
		// Pf("setMinMax: %s\n", t.Interval())
		t.parent.setMinMax()
	}
}
