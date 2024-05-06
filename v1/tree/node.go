package tree

import (
	"time"

	"github.com/reugn/async"
	"github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// Node represents a node in an interval tree.
type Node struct {
	interval interval.Interval
	parent   *Node // Pointer to this node's parent
	left     *Node // Pointer to the left child
	right    *Node // Pointer to the right child

	// minStart is the earliest start time of any interval in the subtree
	// rooted at this node
	minStart time.Time

	// maxEnd is the latest end time of any interval in the subtree
	// rooted at this node
	maxEnd time.Time

	// maxPriority is the highest priority of any interval in the subtree
	// rooted at this node, including this node
	maxPriority float64

	// minPriority is the lowest priority of any interval in the subtree
	// rooted at this node, including this node
	minPriority float64

	// height is the height of the node's subtree, including the node
	height int

	// size is the number of nodes in the node's subtree, including the node
	size int

	mu async.ReentrantLock
}

func (t *Node) MinPriority() float64 {
	return t.minPriority
}

func (t *Node) SetMinPriority(minPriority float64) {
	t.minPriority = minPriority
}

func (t *Node) MaxPriority() float64 {
	return t.maxPriority
}

func (t *Node) SetMaxPriority(maxPriority float64) {
	t.maxPriority = maxPriority
}

func (t *Node) MaxEnd() time.Time {
	return t.maxEnd
}

func (t *Node) SetMaxEnd(maxEnd time.Time) {
	t.maxEnd = maxEnd
}

func (t *Node) MinStart() time.Time {
	return t.minStart
}

func (t *Node) SetMinStart(minStart time.Time) {
	t.minStart = minStart
}

func (t *Node) Right() *Node {
	return t.right
}

func (t *Node) Left() *Node {
	return t.left
}

func (t *Node) Parent() *Node {
	return t.parent
}

func (t *Node) SetParent(parent *Node) {
	t.parent = parent
}

func (t *Node) Height() int {
	return t.height
}

func (t *Node) SetHeight(height int) {
	t.height = height
}

func (t *Node) Size() int {
	return t.size
}

func (t *Node) SetSize(size int) {
	t.size = size
}

// String returns a string representation of the node.
func (t *Node) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := goadapt.Spf("Tree: %p\n", t)
	out += goadapt.Spf("  Interval: %v\n", t.Interval())
	out += goadapt.Spf("  Parent: %p\n", t.parent)
	out += goadapt.Spf("  Left: %p\n", t.left)
	out += goadapt.Spf("  Right: %p\n", t.right)
	out += goadapt.Spf("  MinStart: %v\n", t.minStart)
	out += goadapt.Spf("  MaxEnd: %v\n", t.maxEnd)
	out += goadapt.Spf("  MaxPriority: %v\n", t.maxPriority)
	out += goadapt.Spf("  MinPriority: %v\n", t.minPriority)
	out += goadapt.Spf("  Height: %v\n", t.height)
	out += goadapt.Spf("  Size: %v\n", t.size)
	return out
}

// Busy returns true if the interval is busy.
func (t *Node) Busy() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	goadapt.Assert(t.Interval() != nil, "unexpected nil interval")
	return t.Interval().Busy()
}

// Start returns the start time of the interval in the node.
func (t *Node) Start() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Interval().Start()
}

// End returns the end time of the interval in the node.
func (t *Node) End() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Interval().End()
}

// Interval returns the node's interval.
func (t *Node) Interval() interval.Interval {
	// t.mu.Lock()
	// defer t.mu.Unlock()
	return t.interval
}

// SetInterval sets the node's interval.
func (t *Node) SetInterval(iv interval.Interval) {
	// t.mu.Lock()
	// defer t.mu.Unlock()
	t.interval = iv
}

// newNodeFromInterval creates and returns a new Tree node containing the given interval.
func newNodeFromInterval(interval interval.Interval) *Node {
	return &Node{
		interval:    interval,
		minStart:    interval.Start(),
		maxEnd:      interval.End(),
		minPriority: interval.Priority(),
		maxPriority: interval.Priority(),
		height:      1,
		size:        1,
	}
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
