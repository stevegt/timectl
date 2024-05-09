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
	// path from the root of the tree. The first element
	// of the path is the root node of the tree, and the
	// last element is the node parent
	path Path
	// left child
	left *Node
	// right child
	right *Node

	mu async.ReentrantLock

	nodeCache
}

// clone returns a copy of the node.  We always clone nodes before
// modifying them, so that we don't modify the tree in place.
func (t *Node) clone() *Node {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := &Node{
		interval: t.interval,
		path:     t.path, // XXX get rid of this
		left:     t.left,
		right:    t.right,
	}
	out.clearCache()
	return out
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

// clearCache clears the cache of this node and all its ancestors.
func (t *Node) clearCache() {
	t.nodeCache = nodeCache{}
	if t.Parent() != nil {
		t.Parent().clearCache()
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
	if t == nil {
		return nil
	}
	return t.right
}

func (t *Node) Left() *Node {
	if t == nil {
		return nil
	}
	return t.left
}

func (t *Node) Parent() *Node {
	if t == nil {
		return nil
	}
	if len(t.path) == 0 {
		return nil
	}
	return t.path.Last()
}

func (t *Node) SetParent(parent *Node) (out *Node) {
	if len(t.path) == 0 {
		t.path = append(t.path, parent)
	} else {
		t.path[len(t.path)-1] = parent
	}
	return out
}

func (t *Node) Height() int {
	if t == nil {
		return 0
	}
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
	if t == nil {
		return 0
	}
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
	out += Spf("  Parent: %p\n", t.Parent())
	out += Spf("  Left: %p\n", t.Left())
	out += Spf("  Right: %p\n", t.Right())
	out += Spf("  MinStart: %v\n", t.MinStart())
	out += Spf("  MaxEnd: %v\n", t.MaxEnd())
	out += Spf("  MaxPriority: %v\n", t.MaxPriority())
	out += Spf("  MinPriority: %v\n", t.MinPriority())
	out += Spf("  Height: %v\n", t.Height())
	out += Spf("  Size: %v\n", t.Size())
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
	t.clearCache()
}

// newNodeFromInterval creates and returns a new node containing the given interval.
func newNodeFromInterval(interval interval.Interval) *Node {
	node := &Node{
		interval: interval,
	}
	node.clearCache()
	return node
}

// SetLeft sets the left child of this node.  It returns the old left
// child or nil if there was no old left child.  If the given child node
// is already a child of another node, the right child of this node,
// or the parent of this node, then this function clears the old
// relationship before setting the new one.
// XXX move to Tree
// XXX return new Tree
func (t *Node) SetLeft(left *Node) {
	if left != nil && left.Parent() != nil {
		if left.Parent().left == left {
			left.Parent().left = nil
		}
		if left.Parent().right == left {
			left.Parent().right = nil
		}
	}
	if t.right == left {
		t.right = nil
	}
	t.left = left
	if t.left != nil {
		t.left.SetParent(t)
		t.left.clearCache()
	} else {
		t.clearCache()
	}
	return
}

// SetRight sets the right child of this node.  It returns the old right
// child or nil if there was no old right child.  If the given child node
// is already a child of another node, the left child of this node,
// or the parent of this node, then this function clears the old
// relationship before setting the new one.
// XXX move to Tree
// XXX return new Tree
func (t *Node) SetRight(right *Node) {
	if right != nil && right.Parent() != nil {
		if right.Parent().left == right {
			right.Parent().left = nil
		}
		if right.Parent().right == right {
			right.Parent().right = nil
		}
	}
	if t.left == right {
		t.left = nil
	}
	t.right = right
	if t.right != nil {
		t.right.SetParent(t)
		t.right.clearCache()
	} else {
		t.clearCache()
	}
	return
}

// RotateLeft performs a left rotation on this node.
// XXX move to Tree
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
	R.SetParent(t.Parent())
	t.SetParent(R)
	if R.Parent() != nil {
		switch {
		case R.Parent().left == t:
			R.Parent().left = R
		case R.Parent().right == t:
			R.Parent().right = R
		default:
			Assert(false, "can't find t in R.Parent")
		}
	}
	if x != nil {
		x.SetParent(t)
		x.clearCache()
	} else {
		t.clearCache()
	}
	return
}

// RotateRight performs a right rotation on this node.
// XXX move to Tree
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
	L.SetParent(t.Parent())
	t.SetParent(L)
	if L.Parent() != nil {
		switch {
		case L.Parent().left == t:
			L.Parent().left = L
		case L.Parent().right == t:
			L.Parent().right = L
		default:
			Assert(false, "can't find t in L.Parent")
		}
	}
	if y != nil {
		y.SetParent(t)
		y.clearCache()
	} else {
		t.clearCache()
	}
	return
}

// free sets the interval of the node to a free interval and updates
// the min/max values.  The node's old interval is still intact, but
// no longer part of the tree.  We return the old interval so that the
// caller can decide what to do with it.
// accept Path instead of Node
// XXX return modified tree instead of old interval
func (t *Node) free() (out *Node) {
	out = t
	// XXX should be:
	// out = t.clone()
	freeInterval := interval.NewInterval(out.Start(), out.End(), 0)
	out.SetInterval(freeInterval)
	return
}
