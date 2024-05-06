package tree

import (
	"time"

	"github.com/reugn/async"
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
	"github.com/stevegt/timectl/util"
)

// Node represents a node in an interval tree.
type Node struct {
	interval interval.Interval
	parent   *Node // Pointer to this node's parent
	left     *Node // Pointer to the left child
	right    *Node // Pointer to the right child

	mu async.ReentrantLock

	nodeCache
}

// nodeCache is a cache of selected node fields
type nodeCache struct {
	// minPriority is the lowest priority of any interval in the subtree
	// rooted at this node, including this node
	minPriority *float64

	// maxPriority is the highest priority of any interval in the subtree
	// rooted at this node, including this node
	maxPriority *float64

	// minStart is the earliest start time of any interval in the subtree
	// rooted at this node
	minStart *time.Time

	// maxEnd is the latest end time of any interval in the subtree
	// rooted at this node
	maxEnd *time.Time

	// height is the height of the node's subtree, including the node
	height *int

	// size is the number of nodes in the node's subtree, including the node
	size *int
}

// ClearCache clears the cache of this node and all its ancestors.
func (t *Node) ClearCache() {
	t.nodeCache = nodeCache{}
	if t.parent != nil {
		t.parent.ClearCache()
	}
}

func (t *Node) MinPriority() float64 {
	if t.minPriority != nil {
		return *t.minPriority
	}
	out := t.interval.Priority()
	if t.left != nil {
		out = min(out, t.left.MinPriority())
	}
	if t.right != nil {
		out = min(out, t.right.MinPriority())
	}
	t.minPriority = &out
	return out
}

func (t *Node) MaxPriority() float64 {
	// check the cache first
	if t.maxPriority != nil {
		return *t.maxPriority
	}
	out := t.interval.Priority()
	if t.left != nil {
		out = max(out, t.left.MaxPriority())
	}
	if t.right != nil {
		out = max(out, t.right.MaxPriority())
	}
	t.maxPriority = &out
	return out
}

func (t *Node) MaxEnd() time.Time {
	if t.maxEnd != nil {
		return *t.maxEnd
	}
	out := t.Interval().End()
	if t.right != nil {
		out = util.MaxTime(out, t.right.MaxEnd())
	}
	t.maxEnd = &out
	return out
}

func (t *Node) MinStart() time.Time {
	if t.minStart != nil {
		return *t.minStart
	}
	out := t.Interval().Start()
	if t.left != nil {
		out = util.MinTime(out, t.left.MinStart())
	}
	t.minStart = &out
	return out
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

func (t *Node) Height() int {
	if t.height != nil {
		return *t.height
	}
	var leftHeight, rightHeight int
	if t.left != nil {
		leftHeight = t.left.Height()
	}
	if t.right != nil {
		rightHeight = t.right.Height()
	}
	out := 1 + max(leftHeight, rightHeight)
	t.height = &out
	return out
}

func (t *Node) Size() int {
	if t.size != nil {
		return *t.size
	}
	var leftSize, rightSize int
	if t.left != nil {
		leftSize = t.left.Size()
	}
	if t.right != nil {
		rightSize = t.right.Size()
	}
	out := 1 + leftSize + rightSize
	t.size = &out
	return out
}

// String returns a string representation of the node.
func (t *Node) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := Spf("Tree: %p\n", t)
	out += Spf("  Interval: %v\n", t.Interval())
	out += Spf("  Parent: %p\n", t.parent)
	out += Spf("  Left: %p\n", t.left)
	out += Spf("  Right: %p\n", t.right)
	out += Spf("  MinStart: %v\n", t.minStart)
	out += Spf("  MaxEnd: %v\n", t.maxEnd)
	out += Spf("  MaxPriority: %v\n", t.maxPriority)
	out += Spf("  MinPriority: %v\n", t.minPriority)
	out += Spf("  Height: %v\n", t.height)
	out += Spf("  Size: %v\n", t.size)
	return out
}

// Busy returns true if the interval is busy.
func (t *Node) Busy() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	Assert(t.Interval() != nil, "unexpected nil interval")
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
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.interval
}

// SetInterval sets the node's interval.
func (t *Node) SetInterval(iv interval.Interval) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.interval = iv
	t.ClearCache()
}

// newNodeFromInterval creates and returns a new Tree node containing the given interval.
func newNodeFromInterval(interval interval.Interval) *Node {
	node := &Node{
		interval: interval,
	}
	node.ClearCache()
	return node
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
		t.left.ClearCache()
	} else {
		t.ClearCache()
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
		t.right.ClearCache()
	} else {
		t.ClearCache()
	}
	return
}

// RotateLeft performs a left rotation on this node.
func (t *Node) RotateLeft() (R *Node) {
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
		x.ClearCache()
	} else {
		t.ClearCache()
	}
	return
}

// RotateRight performs a right rotation on this node.
func (t *Node) RotateRight() (L *Node) {
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
		y.ClearCache()
	} else {
		t.ClearCache()
	}
	return
}
