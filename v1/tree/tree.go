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

// Tree represents a node in an interval tree.
type Tree struct {
	Interval interval.Interval
	Parent   *Tree // Pointer to this node's parent
	Left     *Tree // Pointer to the Left child
	Right    *Tree // Pointer to the Right child

	// MinStart is the earliest start time of any Interval in the subtree
	// rooted at this node
	MinStart time.Time

	// MaxEnd is the latest end time of any Interval in the subtree
	// rooted at this node
	MaxEnd time.Time

	// MaxPriority is the highest priority of any Interval in the subtree
	// rooted at this node
	MaxPriority float64

	// Height is the height of the node's subtree, including the node
	Height int

	// Size is the number of nodes in the node's subtree, including the node
	Size int

	Mu async.ReentrantLock
}

// NewTree creates and returns a new Tree node containing a free interval spanning all time.
func NewTree() *Tree {
	return newTreeFromInterval(interval.NewInterval(TreeStart, TreeEnd, 0))
}

// newTreeFromInterval creates and returns a new Tree node containing the given interval.
func newTreeFromInterval(interval interval.Interval) *Tree {
	return &Tree{
		Interval:    interval,
		MinStart:    interval.Start(),
		MaxEnd:      interval.End(),
		MaxPriority: interval.Priority(),
	}
}

// SetLeft sets the Left child of this node.  It returns the old Left
// child or nil if there was no old Left child.
func (t *Tree) SetLeft(left *Tree) (old *Tree) {
	old = t.Left
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
// child or nil if there was no old Right child.
func (t *Tree) SetRight(right *Tree) (old *Tree) {
	old = t.Right
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
func (t *Tree) Insert(newInterval interval.Interval) bool {
	t.Mu.Lock()
	defer t.Mu.Unlock()

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
func (t *Tree) ckHeight() {
	if t == nil {
		return
	}
	calculatedHeight := t.Height
	actualHeight := t.height()
	Assert(calculatedHeight == actualHeight, "height mismatch Height: %d height(): %d", calculatedHeight, actualHeight)
}

func (t *Tree) setMaxPriority() {
	t.MaxPriority = t.Interval.Priority()
	if t.Left != nil {
		t.MaxPriority = max(t.MaxPriority, t.Left.MaxPriority)
	}
	if t.Right != nil {
		t.MaxPriority = max(t.MaxPriority, t.Right.MaxPriority)
	}
}

// BusyIntervals returns a slice of all busy intervals in all leaf nodes of the tree.
func (t *Tree) BusyIntervals() (intervals []interval.Interval) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	for _, i := range t.AllIntervals() {
		if i.Busy() {
			intervals = append(intervals, i)
		}
	}
	return
}

// AllIntervals returns a slice of all intervals in all leaf nodes of the tree.
func (t *Tree) AllIntervals() []interval.Interval {
	t.Mu.Lock()
	defer t.Mu.Unlock()

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
func (t *Tree) Busy() bool {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	Assert(t.Interval != nil, "unexpected nil interval")
	return t.Interval.Busy()
}

// Start returns the start time of the interval in the node.
func (t *Tree) Start() time.Time {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	return t.Interval.Start()
}

// End returns the end time of the interval in the node.
func (t *Tree) End() time.Time {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	return t.Interval.End()
}

// Conflicts returns a slice of intervals in leaf nodes that overlap with the given interval.
// If includeFree is true, then this function returns all intervals that conflict with the given
// interval, otherwise it returns only busy intervals.
func (t *Tree) Conflicts(iv interval.Interval, includeFree bool) []interval.Interval {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	var conflicts []interval.Interval
	if t.Interval != nil && t.Interval.Conflicts(iv, includeFree) {
		conflicts = append(conflicts, t.Interval)
	} else {
		if t.Left != nil {
			conflicts = append(conflicts, t.Left.Conflicts(iv, includeFree)...)
		}
		if t.Right != nil {
			conflicts = append(conflicts, t.Right.Conflicts(iv, includeFree)...)
		}
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
func (t *Tree) FindFree(first bool, minStart, maxEnd time.Time, duration time.Duration) (free interval.Interval) {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	// Pf("FindFree: first: %v minStart: %v maxEnd: %v duration: %v\n", first, minStart, maxEnd, duration)
	// Pf("busy: %v\n", t.Busy())
	if !t.Busy() {
		start := util.MaxTime(minStart, t.MinStart)
		end := util.MinTime(t.MaxEnd, maxEnd)
		sub := subInterval(first, start, end, duration)
		return sub
	}

	var children []*Tree
	var start, end time.Time
	if first {
		children = []*Tree{t.Left, t.Right}
	} else {
		children = []*Tree{t.Right, t.Left}
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
func (t *Tree) FreeIntervals() (intervals []interval.Interval) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	for _, i := range t.AllIntervals() {
		if !i.Busy() {
			intervals = append(intervals, i)
		}
	}
	return
}

// Path is a slice of Tree nodes.
type Path []*Tree

// Append returns a new Path with the given node appended to the end.
func (p Path) Append(t *Tree) Path {
	// because append may reallocate the underlying array, we need to
	// use copy instead of append to avoid modifying the original path
	newPath := make(Path, len(p)+1)
	copy(newPath, p)
	newPath[len(p)] = t
	return newPath
}

// Last returns the last node in the path.
func (p Path) Last() *Tree {
	return p[len(p)-1]
}

// String returns a string representation of the path.
func (p Path) String() string {
	var s string
	var parent *Tree
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
func (t *Tree) allPaths(path Path) (c chan Path) {
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
func (t *Tree) allPathsBlocking(path Path, c chan Path) {
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
func (t *Tree) allNodes(fwd bool, start, end time.Time) <-chan *Tree {
	c := make(chan *Tree)
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
func (t *Tree) allNodesBlocking(fwd bool, start, end time.Time, c chan *Tree) {

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

// firstNode returns the first node in the tree.
func (t *Tree) firstNode() *Tree {
	if t.Left != nil {
		return t.Left.firstNode()
	}
	return t
}

// lastNode returns the last node in the tree.
func (t *Tree) lastNode() *Tree {
	if t.Right != nil {
		return t.Right.lastNode()
	}
	return t
}

// AsDot returns a string representation of the tree in Graphviz DOT
// format without relying on any other Tree methods.
func (t *Tree) AsDot(path Path) string {
	t.Mu.Lock()
	defer t.Mu.Unlock()

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
	}
	if t.Right != nil {
		// get right child's dot representation
		out += t.Right.AsDot(path.Append(t.Right))
		// add edge from this node to right child
		out += fmt.Sprintf("  %s -> %sr [label=%s];\n", id, id, "r")
	}
	if top {
		out += "}\n"
	}
	return out
}

// rotateLeft performs a Left rotation on this node. The new root is
// the current node's Right child.  The current node becomes the new
// root's Left child.
func (t *Tree) rotateLeft() (R *Tree) {
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
	// the new root is the current node's right child
	R = t.Right

	// the current node becomes the new root's left child
	//
	//         R
	//        / \
	//   x   t   y
	//
	// detach t and R from each other
	t.Right = nil
	R.Parent = nil
	// set t as R's left child and save a pointer to x
	x := R.SetLeft(t)

	// set the current node's right child to the new
	// root's old left child
	//
	//         R
	//        / \
	//       t   y
	//		  \
	//		   x
	//
	R.Left.SetRight(x)

	// XXX is this needed?
	R.Right.setMinMax()
	return
}

// rotateRight performs a Right rotation on this node.
func (t *Tree) rotateRight() (L *Tree) {
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
	// the new root is the current node's left child
	L = t.Left

	// the current node becomes the new root's right child
	//
	//     L
	//    / \
	//   x   t   y
	//
	// detach t and L from each other
	t.Left = nil
	L.Parent = nil
	// set t as L's right child and save a pointer to y
	y := L.SetRight(t)

	// finally, we set the current node's left child to the new root's
	// old right child
	//
	//     L
	//    / \
	//   x   t
	//      /
	//     y
	//
	L.Right.SetLeft(y)

	// XXX is this needed?
	L.Left.setMinMax()
	return
}

// setMinMax updates the minimum and maximum values of this node and
// its ancestors.
func (t *Tree) setMinMax() {
	if t == nil {
		return
	}
	var leftHeight, rightHeight int
	var leftSize, rightSize int
	if t.Left == nil {
		t.MinStart = t.Interval.Start()
	} else {
		t.MinStart = t.Left.MinStart
		leftHeight = t.Left.Height
		leftSize = t.Left.Size
	}
	if t.Right == nil {
		t.MaxEnd = t.Interval.End()
	} else {
		t.MaxEnd = t.Right.MaxEnd
		rightHeight = t.Right.Height
		rightSize = t.Right.Size
	}
	t.setMaxPriority()

	// the height of the node is the height of the tallest child plus 1
	t.Height = 1 + max(leftHeight, rightHeight)
	// the size of the node is the size of the left child plus the size
	// of the right child plus 1
	t.Size = 1 + leftSize + rightSize

	if t.Parent != nil {
		t.Parent.setMinMax()
	}
}
