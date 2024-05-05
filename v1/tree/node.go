package tree

import (
	"github.com/reugn/async"
	"github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
	"time"
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
