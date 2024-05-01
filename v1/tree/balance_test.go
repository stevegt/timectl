package tree

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
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

	Verify(t, top, false, false)
}

// test conversion to vine
func TestTreeToVine(t *testing.T) {
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
	var size int
	top, size = top.treeToVine()
	// ShowDot(top, false)

	Tassert(t, size == 10, "should be 10 nodes")
	Tassert(t, len(top.BusyIntervals()) == 8, "should be 8 intervals")
	pathChan := top.allPaths(nil)
	expect := "t"
	for path := range pathChan {
		Tassert(t, path.String() == expect, "path should be %v, got %v", expect, path)
		expect += "r"
	}
}

// test vineToTree
func TestVineToTree(t *testing.T) {
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
	var size int
	top, size = top.treeToVine()
	// ShowDot(top, false)

	Tassert(t, size == 10, "should be 10 nodes, got %v", size)
	Tassert(t, len(top.BusyIntervals()) == 8, "should be 8 intervals")
	pathChan := top.allPaths(nil)
	expect := "t"
	for path := range pathChan {
		Tassert(t, path.String() == expect, "path should be %v, got %v", expect, path)
		expect += "r"
	}

	// convert the vine into a balanced tree using the DSW algorithm
	// and the existing rotateLeft() and rotateRight() functions
	top = top.vineToTree(size)
	// ShowDot(top, false)

	Tassert(t, len(top.BusyIntervals()) == 8, "should be 8 intervals")

	Verify(t, top, true, false)
}

// test rebalance with large random trees
func TestRebalanceRandom(t *testing.T) {
	rand.Seed(1)

	// do a bunch of times
	for round := 0; round < 10; round++ {
		top := NewTree()
		// insert random intervals into the tree
		inserted := 0
		for i := 0; i < 10; i++ {
			startMonth := time.Month(rand.Intn(12) + 1)
			startDay := rand.Intn(31) + 1
			startHour := rand.Intn(24)
			startMinute := rand.Intn(60)
			duration := rand.Intn(600) + 1
			startTime := time.Date(2024, startMonth, startDay, startHour, startMinute, 0, 0, time.UTC)
			endTime := startTime.Add(time.Minute * time.Duration(duration))
			iv := interval.NewInterval(startTime, endTime, 1)
			ok := top.Insert(iv)
			if ok {
				inserted++
			}
		}

		// check the counts
		countAll := len(top.AllIntervals())
		countBusy := len(top.BusyIntervals())
		Tassert(t, countBusy == inserted, "should be %v intervals, got %v", inserted, countBusy)

		// rebalance the tree
		top = top.rebalance()

		// verify the tree
		err := top.Verify(true)
		if err != nil {
			Pf("round %v\n", round)
			Pf("inserted: %v\n", inserted)
			Pf("busy intervals: %v\n", len(top.BusyIntervals()))
			Pf("all intervals: %v\n", len(top.AllIntervals()))
			ShowDot(top, false)
			Tassert(t, false, err)
		}

		// check the counts
		gotCountAll := len(top.AllIntervals())
		gotCountBusy := len(top.BusyIntervals())
		Tassert(t, gotCountBusy == inserted, "should be %v intervals, got %v", inserted, gotCountBusy)
		Tassert(t, gotCountAll == countAll, "should be %v intervals, got %v", countAll, gotCountAll)
	}
}