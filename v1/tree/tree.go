package tree

import (
	"fmt"
	"os/exec"
	"strings"
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

// Insert adds a new interval to the tree, adjusting the structure as
// necessary.  Insertion fails if the new interval conflicts with any
// existing interval in the tree.  The new interval must be busy.
func (t *Tree) Insert(newInterval interval.Interval) bool {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	if !newInterval.Busy() {
		// XXX return a meaningful error
		return false
	}

	if t.Busy() {
		if t.Left != nil && newInterval.Start().Before(t.Left.End()) {
			if t.Left.Insert(newInterval) {
				t.setMinMax()
				return true
			}
		}
		if t.Right != nil && newInterval.End().After(t.Right.Start()) {
			if t.Right.Insert(newInterval) {
				t.MaxEnd = t.Right.MaxEnd
				t.setMaxPriority()
				return true
			}
		}
		return false
	}

	// t is a free node, possibly with free children -- we're going to
	// completely replace it with the results of punching a hole in
	// this node's interval
	newIntervals := t.Interval.Punch(newInterval)
	switch len(newIntervals) {
	case 0:
		// newInterval doesn't fit in this node's interval
		return false
	case 1:
		// newInterval fits exactly in this node's interval
		t.Interval = newInterval
		// clear out any children, because free nodes aren't supposed
		// to have children
		t.Left = nil
		t.Right = nil
		t.setMinMax()
		return true
	case 2:
		// newInterval fits in this node's interval, so we put the
		// first interval in this node and the second interval in a
		// new right child
		t.Interval = newIntervals[0]
		t.Left = nil
		t.Right = newTreeFromInterval(newIntervals[1])
		t.setMinMax()
		return true
	case 3:
		// newInterval fits in this node's interval, so we put the
		// first interval in the left child, the second interval in
		// this node, and the third interval in the right child
		t.Left = newTreeFromInterval(newIntervals[0])
		t.Interval = newIntervals[1]
		t.Right = newTreeFromInterval(newIntervals[2])
		t.setMinMax()
		return true
	default:
		panic("unexpected number of intervals")
	}
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

// Busy returns true if any interval in the tree is busy.
func (t *Tree) Busy() bool {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	if t.Interval != nil && t.Interval.Busy() {
		return true
	}
	if t.Left != nil && t.Left.Busy() {
		return true
	}
	if t.Right != nil && t.Right.Busy() {
		return true
	}
	return false
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
	label := Spf("%v\\nminStart %v\\nmaxEnd %v\\nmaxPriority %v", id, t.MinStart, t.MaxEnd, t.MaxPriority)
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

// showDot displays the tree in xdot.  If bg is true, then the xdot
// window is displayed from a background process.
func showDot(tree *Tree, bg bool) {
	dot := tree.AsDot(nil)
	// call 'xdot -' passing the dot file as input
	cmd := exec.Command("xdot", "-")
	cmd.Stdin = strings.NewReader(dot)
	if bg {
		cmd.Start()
		go cmd.Wait()
		return
	}
	cmd.Run()
}
