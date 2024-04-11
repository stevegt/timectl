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
	tree.Insert(interval2)

	Tassert(t, tree.left != nil, "Expected left child")
	Tassert(t, tree.right != nil, "Expected right child")

	Tassert(t, tree.left.Start().Equal(start1), fmt.Sprintf("Expected start1, got %v", tree.left.Start()))
	Tassert(t, tree.right.Start().Equal(start2), fmt.Sprintf("Expected start2, got %v", tree.right.Start()))
}
