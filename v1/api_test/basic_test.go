package api_test

import (
	"fmt"
	"testing"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/tree"
)

// Test a tree node with children.
// A tree is a tree of busy and free intervals that
// span the entire range from treeStart to treeEnd.
func TestTreeStructure(t *testing.T) {
	top := tree.NewTree()
	// insert interval into empty tree
	err := tree.InsertExpect(top, "", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)
	// the other nodes should be non-busy
	err = tree.Expect(top, "l", tree.TreeStartStr, "2024-01-01T10:00:00Z", 0)
	Tassert(t, err == nil, err)
	err = tree.Expect(top, "r", "2024-01-01T11:00:00Z", tree.TreeEndStr, 0)
	Tassert(t, err == nil, err)

	tree.Verify(t, top, false)
}

func TestInsertMany(t *testing.T) {
	top := tree.NewTree()

	// insert several intervals into the tree
	err := tree.InsertExpect(top, "", "2024-01-01T15:00:00Z", "2024-01-01T16:00:00Z", 1)
	tree.Verify(t, top, true)
	Tassert(t, err == nil, err)
	err = tree.InsertExpect(top, "l", "2024-01-01T08:00:00Z", "2024-01-01T09:00:00Z", 1)
	tree.Verify(t, top, true)
	Tassert(t, err == nil, err)
	err = tree.InsertExpect(top, "lr", "2024-01-01T11:00:00Z", "2024-01-01T12:00:00Z", 1)
	tree.Verify(t, top, true)
	Tassert(t, err == nil, err)
	err = tree.InsertExpect(top, "lrr", "2024-01-01T12:00:00Z", "2024-01-01T13:00:00Z", 1)
	tree.Verify(t, top, true)
	Tassert(t, err == nil, err)
	err = tree.InsertExpect(top, "lrrr", "2024-01-01T13:00:00Z", "2024-01-01T14:00:00Z", 1)
	tree.Verify(t, top, true)
	Tassert(t, err == nil, err)
	err = tree.InsertExpect(top, "lrlr", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	tree.Verify(t, top, true)
	Tassert(t, err == nil, err)
	err = tree.InsertExpect(top, "lrrrr", "2024-01-01T14:00:00Z", "2024-01-01T15:00:00Z", 1)
	tree.Verify(t, top, true)
	Tassert(t, err == nil, err)
	// err = tree.InsertExpect(top, "lrl", "2024-01-01T09:00:00Z", "2024-01-01T10:00:00Z", 1)
	// tree.Verify(t, top, true)
	// Tassert(t, err == nil, err)

}

// TestInsertConflict tests inserting an interval that conflicts with
// an existing interval in the
func TestInsertConflict(t *testing.T) {

	top := tree.NewTree()

	// insert an interval into the tree
	err := tree.InsertExpect(top, "", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)

	// insert a conflicting interval
	interval := tree.Insert(top, "2024-01-01T10:30:00Z", "2024-01-01T11:30:00Z", 1)
	Tassert(t, interval == nil, "Expected nil interval")

	tree.Verify(t, top, false)

}

func TestConflicts(t *testing.T) {
	top := tree.NewTree()

	// insert several intervals into the tree
	i1000_1100 := tree.Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, i1000_1100 != nil, "Failed to insert interval")
	i1130_1200 := tree.Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Tassert(t, i1130_1200 != nil, "Failed to insert interval")
	i0900_0930 := tree.Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)
	Tassert(t, i0900_0930 != nil, "Failed to insert interval")

	// create a new interval that overlaps the first interval
	i1030_1130 := tree.NewInterval("2024-01-01T10:30:00Z", "2024-01-01T11:30:00Z", 1)
	// get conflicts for the new interval
	conflicts := top.Conflicts(i1030_1130, false)
	Tassert(t, len(conflicts) == 1, "Expected 1 conflict, got %d", len(conflicts))
	Tassert(t, conflicts[0].Equal(i1000_1100), fmt.Sprintf("Expected %v, got %v", i1000_1100, conflicts[0]))

	// ensure BusyIntervals() returns all intervals
	intervals := top.BusyIntervals()
	Tassert(t, len(intervals) == 3, "Expected 3 intervals, got %d", len(intervals))
	Tassert(t, intervals[0].Equal(i0900_0930), fmt.Sprintf("Expected %v, got %v", i0900_0930, intervals[0]))
	Tassert(t, intervals[1].Equal(i1000_1100), fmt.Sprintf("Expected %v, got %v", i1000_1100, intervals[1]))
	Tassert(t, intervals[2].Equal(i1130_1200), fmt.Sprintf("Expected %v, got %v", i1130_1200, intervals[2]))

	tree.Verify(t, top, false)

}
