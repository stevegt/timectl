package tree

import (
	"time"

	"github.com/reugn/async"
	. "github.com/stevegt/goadapt"
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

	// dirty is true if the node has been modified and Update() has not
	// been called
	dirty bool

	mu async.ReentrantLock
}

// clone returns a copy of the node.
func (t *Node) clone() *Node {
	if t == nil {
		return nil
	}
	clone := &Node{
		interval:    t.interval,
		parent:      t.parent,
		left:        t.left,
		right:       t.right,
		minStart:    t.minStart,
		maxEnd:      t.maxEnd,
		minPriority: t.minPriority,
		maxPriority: t.maxPriority,
		height:      t.height,
		size:        t.size,
		dirty:       t.dirty,
	}
	return clone
}

func (t *Node) MinPriority() float64 {
	t.update()
	return t.minPriority
}

func (t *Node) MaxPriority() float64 {
	out := t.interval.Priority()
	if t.left != nil {
		out = max(out, t.left.MaxPriority())
	}
	if t.right != nil {
		out = max(out, t.right.MaxPriority())
	}
	return out
}

func (t *Node) MaxEnd() time.Time {
	t.update()
	return t.maxEnd
}

func (t *Node) MinStart() time.Time {
	t.update()
	return t.minStart
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
	t.update()
	return t.height
}

func (t *Node) Size() int {
	t.update()
	return t.size
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
	t.SetDirty()
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
		t.left.SetDirty()
	} else {
		t.SetDirty()
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
		t.right.SetDirty()
	} else {
		t.SetDirty()
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
		x.SetDirty()
	} else {
		t.SetDirty()
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
		y.SetDirty()
	} else {
		t.SetDirty()
	}
	return
}

// update updates the minimum and maximum values of this node and
// its ancestors.
func (t *Node) update() {
	if t == nil {
		return
	}

	if !t.dirty {
		return
	}
	t.dirty = false

	var leftHeight, rightHeight int
	var leftSize, rightSize int
	if t.left == nil {
		t.minStart = t.Interval().Start()
	} else {
		t.minStart = t.left.MinStart()
		leftHeight = t.left.Height()
		leftSize = t.left.Size()
	}
	if t.right == nil {
		t.maxEnd = t.Interval().End()
	} else {
		t.maxEnd = t.right.MaxEnd()
		rightHeight = t.right.Height()
		rightSize = t.right.Size()
	}

	t.maxPriority = t.Interval().Priority()
	t.minPriority = t.Interval().Priority()
	if t.left != nil {
		t.maxPriority = max(t.MaxPriority(), t.left.MaxPriority())
		t.minPriority = min(t.MinPriority(), t.left.MinPriority())
	}
	if t.right != nil {
		t.maxPriority = max(t.MaxPriority(), t.right.MaxPriority())
		t.minPriority = min(t.MinPriority(), t.right.MinPriority())
	}

	// the height of the node is the height of the tallest child plus 1
	t.height = 1 + max(leftHeight, rightHeight)
	// the size of the node is the size of the left child plus the size
	// of the right child plus 1
	t.size = 1 + leftSize + rightSize

	if t.parent != nil {
		t.parent.update()
	}

}

// SetDirty sets the dirty flag on the node and all its ancestors.
func (t *Node) SetDirty() {
	t.dirty = true
	if t.parent != nil {
		t.parent.SetDirty()
	}
}
