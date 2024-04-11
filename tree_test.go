package interval

import (
	"fmt"
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
)

func TestTree(t *testing.T) {
	// Tree is an interval tree that stores intervals and allows for
	// fast lookup of intervals given an interval.
	tree := NewTree()
	start1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	interval1 := NewInterval(start1, end1)
	tree.Insert(interval1)

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

// Ensure that the tree is actually a tree node with children
func TestTreeStructure(t *testing.T) {
	tree := NewTree()
	// insert interval into the root node
	start1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end1, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	interval1 := NewInterval(start1, end1)
	tree.Insert(interval1)

	// insert a right interval -- this should cause the root interval
	// to move to the left child and the right interval to move to the
	// right child, and the root interval to be replaced with a new interval
	// that spans the two children
	start2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:30:00")
	Ck(err)
	end2, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T12:30:00")
	Ck(err)
	interval2 := NewInterval(start2, end2)
	tree.Insert(interval2)

	Tassert(t, tree.interval != nil, "Expected root interval")
	Tassert(t, tree.left != nil, "Expected left child")
	Tassert(t, tree.right != nil, "Expected right child")

	Tassert(t, tree.interval.Start().Equal(start1), fmt.Sprintf("Expected start1, got %v", tree.interval.Start()))
	Tassert(t, tree.interval.End().Equal(end2), fmt.Sprintf("Expected end2, got %v", tree.interval.End()))
	Tassert(t, tree.left.interval.Start().Equal(start1), fmt.Sprintf("Expected start1, got %v", tree.left.interval.Start()))
	Tassert(t, tree.right.interval.Start().Equal(start2), fmt.Sprintf("Expected start2, got %v", tree.right.interval.Start()))

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

}
