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

/*
// Tree is a binary tree of nodes.  Each node in the tree contains an
// interval.  Trees are copy-on-write, so any modification to a tree
// returns a new tree.
type Tree struct {
	root *Node
}

// clone returns a copy of the tree.
func (t *Tree) clone() *Tree {
	return &Tree{root: t.root}
}

// NewNewTree creates and returns a new Tree node containing a free interval spanning all time.
func NewNewTree() *Tree {
	out := &Tree{
		root: newNodeFromInterval(interval.NewInterval(TreeStart, TreeEnd, 0)),
	}
	return out
}
*/

// NewTree creates and returns a new Tree node containing a free interval spanning all time.
func NewTree() *Node {
	return newNodeFromInterval(interval.NewInterval(TreeStart, TreeEnd, 0))
}

// Insert clones the tree, adds the interval, and returns the new
// tree. Insertion fails if the new interval conflicts with any
// existing interval in the tree with a priority greater than 0.
// Insertion fails if the new interval is not busy.
// XXX return tree
func (t *Node) Insert(newInterval interval.Interval) (out *Node, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	out = t.clone()

	if !newInterval.Busy() {
		err = fmt.Errorf("new interval is not busy")
		return
	}

	// use FindLowerPriority to find a free interval where we can
	// insert the new interval
	// XXX use subtree instead of nodes
	nodes, _ := t.FindLowerPriority(true, newInterval.Start(), newInterval.End(), newInterval.Duration(), 1)
	// XXX should be something like:
	// nodes, _ := out.FindLowerPriority(true, newInterval.Start(), newInterval.End(), newInterval.Duration(), 1)
	if len(nodes) == 0 {
		err = fmt.Errorf("no free interval found")
		return
	}

	// nodes should be a slice of length 1
	Assert(len(nodes) == 1, "unexpected number of nodes")
	f := nodes[0]
	// freeNode should have a free interval
	Assert(!f.Busy(), "freeNode is busy")

	// f should start on or before newInterval and end on or after
	// newInterval
	if f.Start().After(newInterval.Start()) || f.End().Before(newInterval.End()) {
		err = fmt.Errorf("free interval doesn't wrap new interval")
		return
	}

	newIntervals := f.Interval().Punch(newInterval)
	switch len(newIntervals) {
	case 0:
		// newInterval doesn't fit in this node's interval
		Assert(false, "newInterval doesn't fit in freeNode's interval")
	case 1:
		// newInterval fits exactly in this node's interval
		f.SetInterval(newInterval)
		// XXX return tree
		return nil, nil
	case 2:
		// newInterval fits in this node's interval with a free interval
		// left over
		// put the first interval in this node
		f.SetInterval(newIntervals[0])
		// create a new right child for the second interval and make
		// the old right child the right child of it
		newNode := newNodeFromInterval(newIntervals[1])
		oldRight := f.Right()
		f.SetRight(newNode)
		f.Right().SetRight(oldRight)
		// XXX return tree
		return nil, nil
	case 3:
		// newInterval fits in this node's interval with free intervals
		// remaining to the left and right, so...

		// put the first interval in a new left child, moving the old
		// left child to the left of the new left child
		newLeftNode := newNodeFromInterval(newIntervals[0])
		oldLeft := f.Left()
		f.SetLeft(newLeftNode)
		f.Left().SetLeft(oldLeft)

		// put the second interval in this node
		f.SetInterval(newIntervals[1])

		// put the third interval in a new right child, moving the old
		// right child to the right of the new right child
		newRightNode := newNodeFromInterval(newIntervals[2])
		oldRight := f.SetRight(newRightNode)
		f.Right().SetRight(oldRight)
		// XXX return tree
		return nil, nil
	default:
		Assert(false, "unexpected number of intervals")
	}

	// check everything
	// XXX either remove this or refactor all of the above to use
	// Height in the first place and not need rebalancing
	f.ckHeight()
	f.Right().ckHeight()
	f.Left().ckHeight()

	err = fmt.Errorf("unexpected code path")
	return
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
	if t.Left() != nil {
		intervals = append(intervals, t.Left().AllIntervals()...)
	}
	intervals = append(intervals, t.Interval())
	if t.Right() != nil {
		intervals = append(intervals, t.Right().AllIntervals()...)
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
	if t.Left() != nil {
		conflicts = append(conflicts, t.Left().Conflicts(iv, includeFree)...)
	}
	if t.Right() != nil {
		conflicts = append(conflicts, t.Right().Conflicts(iv, includeFree)...)
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
		start := util.MaxTime(minStart, t.MinStart())
		end := util.MinTime(t.MaxEnd(), maxEnd)
		sub := subInterval(first, start, end, duration)
		return sub
	}

	var children []*Node
	var start, end time.Time
	if first {
		children = []*Node{t.Left(), t.Right()}
	} else {
		children = []*Node{t.Right(), t.Left()}
	}

	for _, child := range children {
		if child == nil {
			continue
		}
		start = util.MaxTime(minStart, child.MinStart())
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
	if t.Left() != nil {
		t.Left().allPathsBlocking(myPath, c)
	}
	c <- myPath
	if t.Right() != nil {
		t.Right().allPathsBlocking(myPath, c)
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
	if t.MinStart().After(end) {
		return
	}

	if fwd {
		t.Left().allNodesBlocking(fwd, start, end, c)
		c <- t
		t.Right().allNodesBlocking(fwd, start, end, c)
	} else {
		t.Right().allNodesBlocking(fwd, start, end, c)
		c <- t
		t.Left().allNodesBlocking(fwd, start, end, c)
	}
}

// FirstNode returns the first node in the tree.
func (t *Node) FirstNode() *Node {
	if t == nil {
		return nil
	}
	if t.Left() != nil {
		return t.Left().FirstNode()
	}
	return t
}

// LastNode returns the last node in the tree.
func (t *Node) LastNode() *Node {
	if t == nil {
		return nil
	}
	if t.Right() != nil {
		return t.Right().LastNode()
	}
	return t
}

// AsDot returns a string representation of the tree in Graphviz DOT format
func (t *Node) AsDot(path Path) string {
	// t.Mu.Lock()
	// defer t.Mu.Unlock()

	if t == nil {
		return ""
	}

	var out string
	var top bool
	if path == nil {
		top = true
		path = Path{t}
		out += "digraph G {\n"
	}
	id := path.String()
	label := Spf("left %p    right %p\\n", t.Left(), t.Right())
	label += Spf("%v\\nminStart %v\\nmaxEnd %v\\nmaxPriority %v", id, t.MinStart(), t.MaxEnd(), t.MaxPriority())
	if t.Interval() != nil {
		label += fmt.Sprintf("\\n%s", t.Interval())
	} else {
		label += "\\nnil"
	}
	out += fmt.Sprintf("  %s [label=\"%s\"];\n", id, label)
	if t.Left() != nil {
		// get left child's dot representation
		out += t.Left().AsDot(path.Append(t.Left()))
		// add edge from this node to left child
		out += fmt.Sprintf("  %s -> %sl [label=%s];\n", id, id, "l")
	} else {
		out += fmt.Sprintf("  %sl [label=\"nil\" style=dotted];\n", id)
		out += fmt.Sprintf("  %s -> %sl [label=%s];\n", id, id, "l")
	}
	if t.Right() != nil {
		// get right child's dot representation
		out += t.Right().AsDot(path.Append(t.Right()))
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
