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

	// showDot(tree, false)

	Verify(t, top, false)
}

// test rebalancing the tree
func TestRebalanceSimple(t *testing.T) {
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

	// insert another interval into the tree
	err = InsertExpect(top, "r", "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Tassert(t, err == nil, err)

	showDot(top, false)

	err = top.ckBalance(nil)
	Tassert(t, err != nil, "tree should be unbalanced")

	// rebalance the tree
	top = top.rebalance()

	Verify(t, top, false)

}
