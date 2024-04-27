package timectl

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
	// . "github.com/stevegt/goadapt"
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
	interval Interval
	left     *Tree // Pointer to the left child
	right    *Tree // Pointer to the right child

	// maxEnd is the latest end time of any interval in the subtree
	// rooted at this node
	maxEnd time.Time

	mu sync.RWMutex
}

// NewTree creates and returns a new Tree node containing a free interval spanning all time.
func NewTree() *Tree {
	return &Tree{
		interval: NewInterval(TreeStart, TreeEnd, 0),
	}
}

// Insert adds a new interval to the tree, adjusting the structure as
// necessary.  Insertion fails if the new interval conflicts with any
// existing interval in the tree.  The new interval must be busy.
func (t *Tree) Insert(newInterval Interval) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.insert(newInterval)
}

// insert is a non-threadsafe version of Insert for internal use.
func (t *Tree) insert(newInterval Interval) (ok bool) {

	if !newInterval.Busy() {
		// XXX return a meaningful error
		return false
	}

	if t.busy() {
		if t.left != nil && newInterval.Start().Before(t.left.End()) {
			if t.left.insert(newInterval) {
				return true
			}
		}
		if t.right != nil && newInterval.End().After(t.right.Start()) {
			if t.right.insert(newInterval) {
				t.maxEnd = t.right.maxEnd
				return true
			}
		}
		return false
	}

	// t is a free node, possibly with free children -- we're going to
	// completely replace it with the results of punching a hole in
	// this node's interval
	newIntervals := t.interval.Punch(newInterval)
	switch len(newIntervals) {
	case 0:
		// newInterval doesn't fit in this node's interval
		return false
	case 1:
		// newInterval fits exactly in this node's interval
		t.interval = newInterval
		// clear out any children, because free nodes aren't supposed
		// to have children
		t.left = nil
		t.right = nil
		t.maxEnd = newInterval.End()
		return true
	case 2:
		// newInterval fits in this node's interval, so we put the
		// first interval in this node and the second interval in a
		// new right child
		t.interval = newIntervals[0]
		t.left = nil
		t.right = &Tree{interval: newIntervals[1]}
		t.maxEnd = t.right.maxEnd
		return true
	case 3:
		// newInterval fits in this node's interval, so we put the
		// first interval in the left child, the second interval in
		// this node, and the third interval in the right child
		t.left = &Tree{interval: newIntervals[0]}
		t.interval = newIntervals[1]
		t.right = &Tree{interval: newIntervals[2]}
		t.maxEnd = t.right.maxEnd
		return true
	default:
		panic("unexpected number of intervals")
	}
}

// MinTime returns the earlier of two time.Time values.
func MinTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// MaxTime returns the later of two time.Time values.
func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// maxDuration returns the longer of two time.Duration values.
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// BusyIntervals returns a slice of all busy intervals in all leaf nodes of the tree.
func (t *Tree) BusyIntervals() (intervals []Interval) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, i := range t.allIntervals() {
		if i.Busy() {
			intervals = append(intervals, i)
		}
	}
	return
}

// AllIntervals returns a slice of all intervals in all leaf nodes of the tree.
func (t *Tree) AllIntervals() []Interval {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.allIntervals()
}

// allIntervals is a non-threadsafe version of AllIntervals for internal use.
func (t *Tree) allIntervals() []Interval {
	var intervals []Interval
	if t.left != nil {
		intervals = append(intervals, t.left.AllIntervals()...)
	}
	intervals = append(intervals, t.interval)
	if t.right != nil {
		intervals = append(intervals, t.right.AllIntervals()...)
	}
	return intervals
}

// Busy returns true if any interval in the tree is busy.
func (t *Tree) Busy() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.busy()
}

// busy is a non-threadsafe version of Busy for internal use.
func (t *Tree) busy() bool {
	if t.interval != nil && t.interval.Busy() {
		return true
	}
	if t.left != nil && t.left.Busy() {
		return true
	}
	if t.right != nil && t.right.Busy() {
		return true
	}
	return false
}

// Start returns the start time of the interval in the node.
func (t *Tree) Start() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.interval.Start()
}

// End returns the end time of the interval in the node.
func (t *Tree) End() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.interval.End()
}

// treeStart returns the start time of the leftmost child node.
func (t *Tree) treeStart() time.Time {
	if t.left == nil {
		return t.interval.Start()
	}
	return t.left.treeStart()
}

// treeEnd returns the end time of the rightmost child node.
func (t *Tree) treeEnd() time.Time {
	if t.right == nil {
		return t.interval.End()
	}
	return t.right.treeEnd()
}

// Conflicts returns a slice of intervals in leaf nodes that overlap with the given interval.
// If includeFree is true, then this function returns all intervals that conflict with the given
// interval, otherwise it returns only busy intervals.
func (t *Tree) Conflicts(interval Interval, includeFree bool) []Interval {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.conflicts(interval, includeFree)
}

// conflicts is a non-threadsafe version of Conflicts for internal
// use.  If includeFree is true, then this function returns all
// intervals that conflict with the given interval, otherwise it
// returns only busy intervals.
func (t *Tree) conflicts(interval Interval, includeFree bool) []Interval {
	var conflicts []Interval
	if t.interval != nil && t.interval.Conflicts(interval, includeFree) {
		conflicts = append(conflicts, t.interval)
	} else {
		if t.left != nil {
			conflicts = append(conflicts, t.left.conflicts(interval, includeFree)...)
		}
		if t.right != nil {
			conflicts = append(conflicts, t.right.conflicts(interval, includeFree)...)
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
func (t *Tree) FindFree(first bool, minStart, maxEnd time.Time, duration time.Duration) (free Interval) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Pf("FindFree: first: %v minStart: %v maxEnd: %v duration: %v\n", first, minStart, maxEnd, duration)
	// Pf("busy: %v\n", t.Busy())
	if !t.Busy() {
		start := MaxTime(minStart, t.treeStart())
		end := MinTime(t.treeEnd(), maxEnd)
		sub := subInterval(first, start, end, duration)
		// Pf("sub: %v\n", sub)
		return sub
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
		start = MaxTime(minStart, child.treeStart())
		end = MinTime(child.treeEnd(), maxEnd)
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
func subInterval(first bool, minStart, maxEnd time.Time, duration time.Duration) Interval {
	if maxEnd.Sub(minStart) < duration {
		return nil
	}
	if first {
		return NewInterval(minStart, minStart.Add(duration), 0)
	}
	return NewInterval(maxEnd.Add(-duration), maxEnd, 0)
}

// dump is a helper function that prints the tree structure to
// stdout.
func dump(tree *Tree, path string) {
	// fmt.Printf("maxGap: %v interval: %v\n", tree.maxGap, tree.interval)
	fmt.Printf("%-10v: %v\n", path, tree.interval)
	if tree.left != nil {
		dump(tree.left, path+"l")
	}
	if tree.right != nil {
		dump(tree.right, path+"r")
	}
}

// FreeIntervals returns a slice of all free intervals in all leaf nodes of the tree.
func (t *Tree) FreeIntervals() (intervals []Interval) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, i := range t.allIntervals() {
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
			if t == parent.left {
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
// The paths are sorted in depth-first order, left child first.
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
// in depth-first order, left child first.
func (t *Tree) allPathsBlocking(path Path, c chan Path) {
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

// firstNode returns the first node in the tree.
func (t *Tree) firstNode() *Tree {
	if t.left != nil {
		return t.left.firstNode()
	}
	return t
}

// lastNode returns the last node in the tree.
func (t *Tree) lastNode() *Tree {
	if t.right != nil {
		return t.right.lastNode()
	}
	return t
}

// AsDot returns a string representation of the tree in Graphviz DOT
// format without relying on any other Tree methods.
func (t *Tree) AsDot(path Path) string {
	var out string
	var top bool
	if path == nil {
		top = true
		path = Path{t}
		out += "digraph G {\n"
	}
	id := path.String()
	label := id
	if t.interval != nil {
		label += fmt.Sprintf("\\n%s", t.interval)
	}
	out += fmt.Sprintf("  %s [label=\"%s\"];\n", id, label)
	if t.left != nil {
		// get left child's dot representation
		out += t.left.AsDot(path.Append(t.left))
		// add edge from this node to left child
		out += fmt.Sprintf("  %s -> %sl [label=%s];\n", id, id, "l")
	}
	if t.right != nil {
		// get right child's dot representation
		out += t.right.AsDot(path.Append(t.right))
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
