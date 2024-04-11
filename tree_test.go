package interval

import (
	"fmt"
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
)

// Tree is an interval tree that stores intervals and allows for fast
// lookup of intervals given an interval.  All inserted intervals are
// stored in leaf nodes; internal nodes store intervals that span the
// intervals of their children.  This allows for fast lookup of
// intervals that overlap a given interval.  Upon insertion of an
// interval, an existing leaf node is split into two leaf nodes, the
// existing interval is moved to one of the two leaf nodes, and the
// new interval is inserted into the other leaf node.  Left node
// intervals always have start times less than or equal to the right
// node interval start times.  If the start times are equal, the left
// node interval end time is less than or equal to the right node
// interval end time.

// expect is a test helper function that checks if the given tree
// node's interval has the expected start and end times. pathStr is
// the path to the node in the tree, where 'l' means to go left and
// 'r' means to go right.  An empty pathStr means to check the root
// node.
func expect(tree *Tree, pathStr, startStr, endStr string) error {
	start, err := time.Parse("2006-01-02T15:04:05", startStr)
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", endStr)
	Ck(err)
	path := []rune(pathStr)
	if len(path) == 0 {
		if tree.interval == nil {
			return fmt.Errorf("Expected non-nil interval")
		}
		if !tree.interval.Start().Equal(start) {
			return fmt.Errorf("Expected %v, got %v", start, tree.interval.Start())
		}
		if !tree.interval.End().Equal(end) {
			return fmt.Errorf("Expected %v, got %v", end, tree.interval.End())
		}
	} else {
		switch path[0] {
		case 'l':
			if tree.left == nil {
				return fmt.Errorf("Expected non-nil left child")
			}
			err := expect(tree.left, string(path[1:]), startStr, endStr)
			if err != nil {
				return fmt.Errorf("%v:%v", path[0], err)
			}
		case 'r':
			if tree.right == nil {
				return fmt.Errorf("Expected non-nil right child")
			}
			err := expect(tree.right, string(path[1:]), startStr, endStr)
			if err != nil {
				return fmt.Errorf("%v:%v", path[0], err)
			}
		default:
			return fmt.Errorf("Invalid path character: %v", path[0])
		}
	}
	return nil
}

// insertExpect is a test helper function that inserts an interval
// into the tree and checks if the tree has the expected structure.
func insertExpect(tree *Tree, pathStr, startStr, endStr string) error {
	start, err := time.Parse("2006-01-02T15:04:05", startStr)
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", endStr)
	Ck(err)
	interval := NewInterval(start, end)
	tree.Insert(interval)
	return expect(tree, pathStr, startStr, endStr)
}

// Test a tree node with children
func TestTreeStructure(t *testing.T) {

	tree := NewTree()
	// insert interval into the root node
	err := insertExpect(tree, "", "2024-01-01T10:00:00", "2024-01-01T11:00:00")
	Tassert(t, err == nil, err)

	// insert a right interval -- this should cause the new interval
	// to be inserted into the right child
	err = insertExpect(tree, "r", "2024-01-01T12:00:00", "2024-01-01T13:00:00")
	Tassert(t, err == nil, err)
	// the root interval should move to the left child
	err = expect(tree, "l", "2024-01-01T10:00:00", "2024-01-01T11:00:00")
	Tassert(t, err == nil, err)
	// the root interval should be replaced with a new interval
	// that spans the two children
	err = expect(tree, "", "2024-01-01T10:00:00", "2024-01-01T13:00:00")
	Tassert(t, err == nil, err)

	// insert a interval earlier than all other intervals -- this should
	// cause the new interval to insert at tree.left.left
	err = insertExpect(tree, "ll", "2024-01-01T09:00:00", "2024-01-01T09:30:00")
	Tassert(t, err == nil, err)
	// tree.left should move to tree.left.right.
	err = expect(tree, "lr", "2024-01-01T10:00:00", "2024-01-01T11:00:00")
	Tassert(t, err == nil, err)
	// tree.left should span tree.left.left and tree.left.right
	err = expect(tree, "l", "2024-01-01T09:00:00", "2024-01-01T11:00:00")
	Tassert(t, err == nil, err)
	// tree should span tree.left and tree.right
	err = expect(tree, "", "2024-01-01T09:00:00", "2024-01-01T13:00:00")

	/*
		// insert an interval between the root and the right child -- this should
		// cause the right child to move to tree.right.right and the new interval
		// to insert at tree.right.left.  The root interval will not
		// change, and tree.right will be replaced with a new interval that spans
		// tree.right.left and tree.right.right.
		start3, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
		Ck(err)
		end3, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:30:00")
		Ck(err)
		interval3 := NewInterval(start3, end3)
		tree.Insert(interval3)
		Tassert(t, tree.interval != nil, "Expected root interval")
		Tassert(t, tree.left != nil, "Expected left child")
		Tassert(t, tree.right != nil, "Expected right child")
		Tassert(t, tree.right.left != nil, "Expected right.left child")
		Tassert(t, tree.right.right != nil, "Expected right.right child")
		Tassert(t, tree.interval.Start().Equal(start1), fmt.Sprintf("Expected start1, got %v", tree.interval.Start()))
		Tassert(t, tree.interval.End().Equal(end2), fmt.Sprintf("Expected end2, got %v", tree.interval.End()))
		Tassert(t, tree.left.interval.Start().Equal(start1), fmt.Sprintf("Expected start1, got %v", tree.left.interval.Start()))
		Tassert(t, tree.right.interval.Start().Equal(start3), fmt.Sprintf("Expected start3, got %v", tree.right.interval.Start()))
		Tassert(t, tree.right.interval.End().Equal(end2), fmt.Sprintf("Expected end2, got %v", tree.right.interval.End()))
		Tassert(t, tree.right.left.interval.Start().Equal(start3), fmt.Sprintf("Expected start3, got %v", tree.right.left.interval.Start()))
		Tassert(t, tree.right.left.interval.End().Equal(end3), fmt.Sprintf("Expected end3, got %v", tree.right.left.interval.End()))
		Tassert(t, tree.right.right.interval.Start().Equal(start2), fmt.Sprintf("Expected start2, got %v", tree.right.right.interval.Start()))
		Tassert(t, tree.right.right.interval.End().Equal(end2), fmt.Sprintf("Expected end2, got %v", tree.right.right.interval.End()))
	*/

}

func TestConflicts(t *testing.T) {
	tree := NewTree()

	// insert an interval into the tree
	start1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	interval1 := NewInterval(start1, end1)
	tree.Insert(interval1)

	// create a new interval that overlaps the first interval
	start2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:30:00")
	Ck(err)
	end2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:30:00")
	Ck(err)
	interval2 := NewInterval(start2, end2)

	// Conflicts returns a slice of Intervals
	intervals := tree.Conflicts(interval2)
	Tassert(t, len(intervals) == 1, fmt.Sprintf("Expected 1 conflict, got %d", len(intervals)))
	Tassert(t, intervals[0].Start().Equal(start1), fmt.Sprintf("Expected start1, got %v", intervals[0].Start()))
	Tassert(t, intervals[0].End().Equal(end1), fmt.Sprintf("Expected end1, got %v", intervals[0].End()))
}
