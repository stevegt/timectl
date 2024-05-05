package tree

import (
	"fmt"
	"time"

	"github.com/reugn/async"

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

// Node represents a node in an interval tree.
type Node struct {
	Interval interval.Interval
	Parent   *Node // Pointer to this node's parent
	Left     *Node // Pointer to the Left child
	Right    *Node // Pointer to the Right child

	// MinStart is the earliest start time of any Interval in the subtree
	// rooted at this node
	MinStart time.Time

	// MaxEnd is the latest end time of any Interval in the subtree
	// rooted at this node
	MaxEnd time.Time

	// MaxPriority is the highest priority of any Interval in the subtree
	// rooted at this node, including this node
	MaxPriority float64

	// MinPriority is the lowest priority of any Interval in the subtree
	// rooted at this node, including this node
	MinPriority float64

	// height is the height of the node's subtree, including the node
	height int

	// size is the number of nodes in the node's subtree, including the node
	size int

	mu async.ReentrantLock
}

// String returns a string representation of the node.
func (t *Node) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := Spf("Tree: %p\n", t)
	out += Spf("  Interval: %v\n", t.Interval)
	out += Spf("  Parent: %p\n", t.Parent)
	out += Spf("  Left: %p\n", t.Left)
	out += Spf("  Right: %p\n", t.Right)
	out += Spf("  MinStart: %v\n", t.MinStart)
	out += Spf("  MaxEnd: %v\n", t.MaxEnd)
	out += Spf("  MaxPriority: %v\n", t.MaxPriority)
	out += Spf("  MinPriority: %v\n", t.MinPriority)
	out += Spf("  Height: %v\n", t.height)
	out += Spf("  Size: %v\n", t.size)
	return out
}

// NewTree creates and returns a new Tree node containing a free interval spanning all time.
func NewTree() *Node {
	return newTreeFromInterval(interval.NewInterval(TreeStart, TreeEnd, 0))
}

// newTreeFromInterval creates and returns a new Tree node containing the given interval.
func newTreeFromInterval(interval interval.Interval) *Node {
	return &Node{
		Interval:    interval,
		MinStart:    interval.Start(),
		MaxEnd:      interval.End(),
		MaxPriority: interval.Priority(),
	}
}

// SetLeft sets the Left child of this node.  It returns the old Left
// child or nil if there was no old Left child.  If the given child node
// is already a child of another node, the Right child of this node,
// or the Parent of this node, then this function clears the old
// relationship before setting the new one.
func (t *Node) SetLeft(left *Node) (old *Node) {
	old = t.Left
	if left != nil && left.Parent != nil {
		if left.Parent.Left == left {
			left.Parent.Left = nil
		}
		if left.Parent.Right == left {
			left.Parent.Right = nil
		}
	}
	if t.Right == left {
		t.Right = nil
	}
	t.Left = left
	if t.Left != nil {
		t.Left.Parent = t
		t.Left.setMinMax()
	} else {
		t.setMinMax()
	}
	return
}

// SetRight sets the Right child of this node.  It returns the old Right
// child or nil if there was no old Right child.  If the given child node
// is already a child of another node, the Left child of this node,
// or the Parent of this node, then this function clears the old
// relationship before setting the new one.
func (t *Node) SetRight(right *Node) (old *Node) {
	old = t.Right
	if right != nil && right.Parent != nil {
		if right.Parent.Left == right {
			right.Parent.Left = nil
		}
		if right.Parent.Right == right {
			right.Parent.Right = nil
		}
	}
	if t.Left == right {
		t.Left = nil
	}
	t.Right = right
	if t.Right != nil {
		t.Right.Parent = t
		t.Right.setMinMax()
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

	newIntervals := f.Interval.Punch(newInterval)
	switch len(newIntervals) {
	case 0:
		// newInterval doesn't fit in this node's interval
		Assert(false, "newInterval doesn't fit in freeNode's interval")
	case 1:
		// newInterval fits exactly in this node's interval
		f.Interval = newInterval
		f.setMinMax()
		return true
	case 2:
		// newInterval fits in this node's interval with a free interval
		// left over
		// put the first interval in this node
		f.Interval = newIntervals[0]
		// create a new right child for the second interval and make
		// the old right child the right child of it
		newNode := newTreeFromInterval(newIntervals[1])
		oldRight := f.SetRight(newNode)
		f.Right.SetRight(oldRight)
		return true
	case 3:
		// newInterval fits in this node's interval with free intervals
		// remaining to the left and right, so...

		// put the first interval in a new left child, moving the old
		// left child to the left of the new left child
		newLeftNode := newTreeFromInterval(newIntervals[0])
		oldLeft := f.SetLeft(newLeftNode)
		f.Left.SetLeft(oldLeft)

		// put the second interval in this node
		f.Interval = newIntervals[1]

		// put the third interval in a new right child, moving the old
		// right child to the right of the new right child
		newRightNode := newTreeFromInterval(newIntervals[2])
		oldRight := f.SetRight(newRightNode)
		f.Right.SetRight(oldRight)
		return true
	default:
		Assert(false, "unexpected number of intervals")
	}

	// check everything
	// XXX either remove this or refactor all of the above to use
	// Height in the first place and not need rebalancing
	f.ckHeight()
	f.Right.ckHeight()
	f.Left.ckHeight()

	Assert(false, "unexpected code path")
	return false
}

// ckHeight checks the calculated height of the node against the actual
// height of the node's subtree.
func (t *Node) ckHeight() {
	if t == nil {
		return
	}
	calculatedHeight := t.height
	actualHeight := t.GetHeight()
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
	if t.Left != nil {
		intervals = append(intervals, t.Left.AllIntervals()...)
	}
	intervals = append(intervals, t.Interval)
	if t.Right != nil {
		intervals = append(intervals, t.Right.AllIntervals()...)
	}
	return intervals
}

// Busy returns true if the interval is busy.
func (t *Node) Busy() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	Assert(t.Interval != nil, "unexpected nil interval")
	return t.Interval.Busy()
}

// Start returns the start time of the interval in the node.
func (t *Node) Start() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Interval.Start()
}

// End returns the end time of the interval in the node.
func (t *Node) End() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Interval.End()
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
	if t.Interval.Conflicts(iv, includeFree) {
		conflicts = append(conflicts, t.Interval)
	}
	if t.Left != nil {
		conflicts = append(conflicts, t.Left.Conflicts(iv, includeFree)...)
	}
	if t.Right != nil {
		conflicts = append(conflicts, t.Right.Conflicts(iv, includeFree)...)
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
		start := util.MaxTime(minStart, t.MinStart)
		end := util.MinTime(t.MaxEnd, maxEnd)
		sub := subInterval(first, start, end, duration)
		return sub
	}

	var children []*Node
	var start, end time.Time
	if first {
		children = []*Node{t.Left, t.Right}
	} else {
		children = []*Node{t.Right, t.Left}
	}

	for _, child := range children {
		if child == nil {
			continue
		}
		start = util.MaxTime(minStart, child.MinStart)
		end = util.MinTime(child.MaxEnd, maxEnd)
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

// Path is a slice of Tree nodes.
type Path []*Node

// Append returns a new Path with the given node appended to the end.
func (p Path) Append(t *Node) Path {
	// because append may reallocate the underlying array, we need to
	// use copy instead of append to avoid modifying the original path
	newPath := make(Path, len(p)+1)
	copy(newPath, p)
	newPath[len(p)] = t
	return newPath
}

// Last returns the last node in the path.
func (p Path) Last() *Node {
	return p[len(p)-1]
}

// String returns a string representation of the path.
func (p Path) String() string {
	var s string
	var parent *Node
	for _, t := range p {
		if parent != nil {
			if t == parent.Left {
				s += "l"
			} else {
				s += "r"
			}
		} else {
			s += "t"
		}
		parent = t
	}
	return s
}

// allPaths returns a channel of all paths to all nodes in the tree.
// The paths are sorted in depth-first order, Left child first.
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
// in depth-first order, Left child first.
func (t *Node) allPathsBlocking(path Path, c chan Path) {
	myPath := path.Append(t)
	// Pf("path %p myPath %p\n", path, myPath)
	// Pf("send: %-10s %v\n", myPath, t.leafInterval)
	if t.Left != nil {
		t.Left.allPathsBlocking(myPath, c)
	}
	c <- myPath
	if t.Right != nil {
		t.Right.allPathsBlocking(myPath, c)
	}
}

// allNodes returns a channel of all nodes in the tree that are
// between the start and end times.  The fwd parameter determines
// whether the nodes are returned in depth-first order, Left child
// first, or in reverse depth-first order, Right child first.
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
// whether the nodes are returned in depth-first order, Left child
// first, or in reverse depth-first order, Right child first.
func (t *Node) allNodesBlocking(fwd bool, start, end time.Time, c chan *Node) {

	if t == nil {
		return
	}

	if t.MaxEnd.Before(start) {
		return
	}
	if t.MinStart.After(end) {
		return
	}

	if fwd {
		t.Left.allNodesBlocking(fwd, start, end, c)
		c <- t
		t.Right.allNodesBlocking(fwd, start, end, c)
	} else {
		t.Right.allNodesBlocking(fwd, start, end, c)
		c <- t
		t.Left.allNodesBlocking(fwd, start, end, c)
	}
}

// FirstNode returns the first node in the tree.
func (t *Node) FirstNode() *Node {
	if t.Left != nil {
		return t.Left.FirstNode()
	}
	return t
}

// LastNode returns the last node in the tree.
func (t *Node) LastNode() *Node {
	if t.Right != nil {
		return t.Right.LastNode()
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
	label := Spf("parent %p\\nthis %p\\n", t.Parent, t)
	label += Spf("left %p    right %p\\n", t.Left, t.Right)
	label += Spf("%v\\nminStart %v\\nmaxEnd %v\\nmaxPriority %v", id, t.MinStart, t.MaxEnd, t.MaxPriority)
	if t.Interval != nil {
		label += fmt.Sprintf("\\n%s", t.Interval)
	} else {
		label += "\\nnil"
	}
	out += fmt.Sprintf("  %s [label=\"%s\"];\n", id, label)
	if t.Left != nil {
		// get left child's dot representation
		out += t.Left.AsDot(path.Append(t.Left))
		// add edge from this node to left child
		out += fmt.Sprintf("  %s -> %sl [label=%s];\n", id, id, "l")
	} else {
		out += fmt.Sprintf("  %sl [label=\"nil\" style=dotted];\n", id)
		out += fmt.Sprintf("  %s -> %sl [label=%s];\n", id, id, "l")
	}
	if t.Right != nil {
		// get right child's dot representation
		out += t.Right.AsDot(path.Append(t.Right))
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

// rotateLeft performs a Left rotation on this node.
func (t *Node) rotateLeft() (R *Node) {
	if t == nil || t.Right == nil {
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
	R = t.Right
	x := R.Left

	// pivot around R
	//
	//         R
	//        / \
	//       t   y
	//		  \
	//		   x
	//
	R.Left = t
	t.Right = x
	R.Parent = t.Parent
	t.Parent = R
	if R.Parent != nil {
		switch {
		case R.Parent.Left == t:
			R.Parent.Left = R
		case R.Parent.Right == t:
			R.Parent.Right = R
		default:
			Assert(false, "can't find t in R.Parent")
		}
	}
	if x != nil {
		x.Parent = t
		x.setMinMax()
	} else {
		t.setMinMax()
	}
	return
}

// rotateRight performs a Right rotation on this node.
func (t *Node) rotateRight() (L *Node) {
	if t == nil || t.Left == nil {
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
	L = t.Left
	y := L.Right

	// pivot around L
	//
	//     L
	//    / \
	//   x   t
	//      /
	//     y
	//
	L.Right = t
	t.Left = y
	L.Parent = t.Parent
	t.Parent = L
	if L.Parent != nil {
		switch {
		case L.Parent.Left == t:
			L.Parent.Left = L
		case L.Parent.Right == t:
			L.Parent.Right = L
		default:
			Assert(false, "can't find t in L.Parent")
		}
	}
	if y != nil {
		y.Parent = t
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
	if t.Left == nil {
		t.MinStart = t.Interval.Start()
	} else {
		t.MinStart = t.Left.MinStart
		leftHeight = t.Left.height
		leftSize = t.Left.size
	}
	if t.Right == nil {
		t.MaxEnd = t.Interval.End()
	} else {
		t.MaxEnd = t.Right.MaxEnd
		rightHeight = t.Right.height
		rightSize = t.Right.size
	}

	t.MaxPriority = t.Interval.Priority()
	t.MinPriority = t.Interval.Priority()
	if t.Left != nil {
		t.MaxPriority = max(t.MaxPriority, t.Left.MaxPriority)
		t.MinPriority = min(t.MinPriority, t.Left.MinPriority)
	}
	if t.Right != nil {
		t.MaxPriority = max(t.MaxPriority, t.Right.MaxPriority)
		t.MinPriority = min(t.MinPriority, t.Right.MinPriority)
	}

	// the height of the node is the height of the tallest child plus 1
	t.height = 1 + max(leftHeight, rightHeight)
	// the size of the node is the size of the left child plus the size
	// of the right child plus 1
	t.size = 1 + leftSize + rightSize

	if t.Parent != nil {
		// Pf("setMinMax: %s\n", t.Interval)
		t.Parent.setMinMax()
	}
}

var seen []*Node
