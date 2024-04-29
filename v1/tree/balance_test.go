package tree

import (
	"testing"

	. "github.com/stevegt/goadapt"
)

// test rotation
func TestRotate(t *testing.T) {
	top := NewTree()

	// insert an interval into the tree
	err := InsertExpect(top, "", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)
	// check the nodes
	err = Expect(top, "l", TreeStartStr, "2024-01-01T10:00:00Z", 0)
	Tassert(t, err == nil, err)
	err = Expect(top, "r", "2024-01-01T11:00:00Z", TreeEndStr, 0)
	Tassert(t, err == nil, err)

	// rotate left
	top = top.rotateLeft()
	// check the nodes
	err = Expect(top, "ll", TreeStartStr, "2024-01-01T10:00:00Z", 0)
	Tassert(t, err == nil, err)
	err = Expect(top, "l", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)
	err = Expect(top, "", "2024-01-01T11:00:00Z", TreeEndStr, 0)
	Tassert(t, err == nil, err)

	// ShowDot(tree, false)

	Verify(t, top, false)
}

// test conversion to vine
func TestVine(t *testing.T) {
	top := NewTree()

	// insert several intervals into the tree
	Insert(top, "2024-01-01T15:00:00Z", "2024-01-01T16:00:00Z", 1)
	Insert(top, "2024-01-01T08:00:00Z", "2024-01-01T09:00:00Z", 1)
	Insert(top, "2024-01-01T11:00:00Z", "2024-01-01T12:00:00Z", 1)
	Insert(top, "2024-01-01T12:00:00Z", "2024-01-01T13:00:00Z", 1)
	Insert(top, "2024-01-01T13:00:00Z", "2024-01-01T14:00:00Z", 1)
	Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Insert(top, "2024-01-01T14:00:00Z", "2024-01-01T15:00:00Z", 1)
	Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T10:00:00Z", 1)

	Tassert(t, len(top.BusyIntervals()) == 8, "should be 8 intervals")

	// convert the tree into a vine
	top = top.treeToVine()
	// ShowDot(top, false)

	Tassert(t, len(top.BusyIntervals()) == 8, "should be 8 intervals")
	pathChan := top.allPaths(nil)
	expect := "t"
	for path := range pathChan {
		Tassert(t, path.String() == expect, "path should be %v, got %v", expect, path)
		expect += "r"
	}
}

// test rebalancing the tree
func XXXTestRebalanceSimple(t *testing.T) {
	top := NewTree()

	// insert 1 interval into the tree
	err := InsertExpect(top, "", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)
	// check the nodes
	err = Expect(top, "l", TreeStartStr, "2024-01-01T10:00:00Z", 0)
	Tassert(t, err == nil, err)
	err = Expect(top, "r", "2024-01-01T11:00:00Z", TreeEndStr, 0)
	Tassert(t, err == nil, err)

	Verify(t, top, false)

	// rebalance is a function that rebalances the tree using the DSW
	// algorithm.  It calls rotateLeft() and rotateRight() to first
	// convert the tree into a vine and then convert the vine into a
	// balanced tree.

	// rebalance the tree
	top.rebalance()
	// nodes should be the same
	err = Expect(top, "l", TreeStartStr, "2024-01-01T10:00:00Z", 0)
	Tassert(t, err == nil, err)
	err = Expect(top, "", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)
	err = Expect(top, "r", "2024-01-01T11:00:00Z", TreeEndStr, 0)
	Tassert(t, err == nil, err)

	Verify(t, top, false)

	// insert more intervals
	err = InsertExpect(top, "r", "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Tassert(t, err == nil, err)
	err = InsertExpect(top, "rr", "2024-01-01T12:30:00Z", "2024-01-01T13:00:00Z", 1)
	Tassert(t, err == nil, err)

	// check the heights
	leftHeight := top.Left.height()
	rightHeight := top.Right.height()
	Tassert(t, leftHeight == 1, "left height should be 1")
	Tassert(t, rightHeight == 3, "right height should be 3")

	err = top.ckBalance(nil)
	Tassert(t, err != nil, "tree should be unbalanced")

	// ShowDot(top, false)

	// rebalance the tree
	top.rebalance()

	// ShowDot(top, false)

	// check the heights
	leftHeight = top.Left.height()
	rightHeight = top.Right.height()
	Tassert(t, leftHeight == 2, "left height should be 2")
	Tassert(t, rightHeight == 2, "right height should be 2")

	Verify(t, top, false)

}

// test rebalancing the tree
func XXXTestRebalance(t *testing.T) {
	top := NewTree()

	// insert a few intervals into the tree
	Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)
	Insert(top, "2024-01-01T14:00:00Z", "2024-01-01T15:00:00Z", 1)

	// rebalance the tree
	top.rebalance()

	err := top.Verify()
	Tassert(t, err == nil, err)

	Verify(t, top, false)

}
