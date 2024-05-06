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
		t.left.SetMinMax()
	} else {
		t.SetMinMax()
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
		t.right.SetMinMax()
	} else {
		t.SetMinMax()
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
		x.SetMinMax()
	} else {
		t.SetMinMax()
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
		y.SetMinMax()
	} else {
		t.SetMinMax()
	}
	return
}

// SetMinMax updates the minimum and maximum values of this node and
// its ancestors.
func (t *Node) SetMinMax() {
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
	if t.left == nil {
		t.SetMinStart(t.Interval().Start())
	} else {
		t.SetMinStart(t.left.MinStart())
		leftHeight = t.left.Height()
		leftSize = t.left.Size()
	}
	if t.right == nil {
		t.SetMaxEnd(t.Interval().End())
	} else {
		t.SetMaxEnd(t.right.MaxEnd())
		rightHeight = t.right.Height()
		rightSize = t.right.Size()
	}

	t.SetMaxPriority(t.Interval().Priority())
	t.SetMinPriority(t.Interval().Priority())
	if t.left != nil {
		t.SetMaxPriority(max(t.MaxPriority(), t.left.MaxPriority()))
		t.SetMinPriority(min(t.MinPriority(), t.left.MinPriority()))
	}
	if t.right != nil {
		t.SetMaxPriority(max(t.MaxPriority(), t.right.MaxPriority()))
		t.SetMinPriority(min(t.MinPriority(), t.right.MinPriority()))
	}

	// the height of the node is the height of the tallest child plus 1
	t.SetHeight(1 + max(leftHeight, rightHeight))
	// the size of the node is the size of the left child plus the size
	// of the right child plus 1
	t.SetSize(1 + leftSize + rightSize)

	if t.parent != nil {
		// Pf("setMinMax: %s\n", t.Interval())
		t.parent.SetMinMax()
	}
}
