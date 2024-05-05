package tree

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// . "github.com/stevegt/goadapt"

// test merging free nodes
func TestMergeFree(t *testing.T) {
	top := NewTree()

	// split the root node into three free nodes
	splitAt1200, err := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
	splitAt1400, err := time.Parse(time.RFC3339, "2024-01-01T14:00:00Z")
	Ck(err)
	top.left = &Node{interval: interval.NewInterval(TreeStart, splitAt1200, 0)}
	top.SetInterval(interval.NewInterval(splitAt1200, splitAt1400, 0))
	top.right = &Node{interval: interval.NewInterval(splitAt1400, TreeEnd, 0)}
	top.left.setMinMax()
	top.right.setMinMax()
	top.setMinMax()

	// Verify(t, top, false, true)

	err = top.Verify(true)
	Tassert(t, err != nil, "Expected error, got nil")

	/*
		// check the tree to vine conversion
		tmp, tmpSize := top.treeToVine()
		Tassert(t, tmp != nil, "Expected non-nil tree")
		Tassert(t, tmpSize == 3, "Expected 3 nodes, got %d", tmpSize)
		ShowDot(tmp, false)
	*/

	// merge the free nodes
	top = top.mergeFree()

	Verify(t, top, false, true)

	// check that the tree has one free interval
	freeIntervals := top.FreeIntervals()
	Tassert(t, len(freeIntervals) == 1, "Expected 1 interval, got %d", len(freeIntervals))
	iv := freeIntervals[0]
	Tassert(t, iv.Start().Equal(TreeStart), fmt.Sprintf("Expected %v, got %v", TreeStart, iv.Start()))
	Tassert(t, iv.End().Equal(TreeEnd), fmt.Sprintf("Expected %v, got %v", TreeEnd, iv.End()))

	// add some busy intervals
	i1000_1100 := Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	i1200_1300 := Insert(top, "2024-01-01T12:00:00Z", "2024-01-01T13:00:00Z", 1)
	i1300_1400 := Insert(top, "2024-01-01T13:00:00Z", "2024-01-01T14:00:00Z", 1)
	_ = i1000_1100
	_ = i1200_1300
	_ = i1300_1400

	Verify(t, top, false, true)

	// ShowDot(top, false)

	// merge the free nodes -- this will reshape the tree but not change the intervals
	top = top.mergeFree()
	// ShowDot(top, false)
	busyIntervals := top.BusyIntervals()
	Tassert(t, len(busyIntervals) == 3, "Expected 3 intervals, got %d", len(busyIntervals))

}

// test mergeFree complex
func TestMergeFreeComplex(t *testing.T) {
	rand.Seed(1)
	top := NewTree()

	var inserted []interval.Interval
	for i := 0; i < 100; i++ {
		// randomly insert or free an interval
		switch rand.Intn(2) {
		case 0:
			// insert an interval
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
				inserted = append(inserted, iv)
			}
		case 1:
			// free a random interval from the inserted intervals
			if len(inserted) > 0 {
				index := rand.Intn(len(inserted))
				iv := inserted[index]
				iv.SetPriority(0)
				// remove the interval from the list
				inserted = append(inserted[:index], inserted[index+1:]...)
			}
		}
		// merge the free nodes
		top = top.mergeFree()
		// ensure that there are no free intervals in the inserted list
		for _, iv := range inserted {
			Tassert(t, iv.Priority() > 0, "Expected busy interval, got free interval")
		}
		// verify the tree
		err := top.Verify(false)
		Tassert(t, err == nil, "i=%d: %v", i, err)
	}
}

/* deletion algorithm:

- Find the exact interval in the tree
- If the exact interval is not found, then return false
- If the exact interval is found in a leaf node, then remove the node and clear the link in the parent
- if either of the parent's children is a leaf node, then promote the child
- starting with the node's grandparent, walk the tree, merging free nodes as necessary

*/

// test free
func TestFree(t *testing.T) {
	top := NewTree()

	// insert an interval into the tree
	iv := NewInterval("2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	ok := top.Insert(iv)
	Tassert(t, ok, "Failed to insert interval")

	// find the node containing the interval
	path, found := top.FindExact(iv)
	_ = path

	// free the node. The free() function replaces the interval in the
	// node with a free interval that spans the same range.  The
	// function does not merge free nodes.
	freed := top.free(found)
	Tassert(t, freed.Equal(iv), fmt.Sprintf("Expected %v, got %v", iv, freed))

	// check that the interval is no longer in the tree
	intervals := top.BusyIntervals()
	Tassert(t, len(intervals) == 0, "Expected 0 intervals, got %d", len(intervals))
	// we haven't merged free nodes yet, so there should be three free nodes
	freeIntervals := top.FreeIntervals()
	Tassert(t, len(freeIntervals) == 3, "Expected 3 free intervals, got %d", len(freeIntervals))

	// merge the free nodes
	top = top.mergeFree()
	// there should now be one free interval
	freeIntervals = top.FreeIntervals()
	Tassert(t, len(freeIntervals) == 1, "Expected 1 interval, got %d", len(freeIntervals))
	freeInterval := freeIntervals[0]
	Tassert(t, freeInterval.Start().Equal(TreeStart), fmt.Sprintf("Expected %v, got %v", TreeStart, freeInterval.Start()))
	Tassert(t, freeInterval.End().Equal(TreeEnd), fmt.Sprintf("Expected %v, got %v", TreeEnd, freeInterval.End()))

	Verify(t, top, false, false)
}

// test delete
func TestDelete(t *testing.T) {
	top := NewTree()

	// insert an interval into the tree
	iv := NewInterval("2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	ok := top.Insert(iv)
	Tassert(t, ok, "Failed to insert interval")

	// find the node containing the interval
	path, found := top.FindExact(iv)
	_ = path

	// delete the node. The Delete() function replaces the interval in the
	// node with a free interval that spans the same range, and then merges
	// free nodes.
	top, err := top.Delete(found.Interval())
	Tassert(t, err == nil, err)

	// check that the interval is no longer in the tree
	intervals := top.BusyIntervals()
	Tassert(t, len(intervals) == 0, "Expected 0 intervals, got %d", len(intervals))
	// we've merged free nodes, so there should be one free node
	freeIntervals := top.FreeIntervals()
	Tassert(t, len(freeIntervals) == 1, "Expected 1 free intervals, got %d", len(freeIntervals))

	Verify(t, top, false, false)

}

// test complex delete
func TestDeleteComplex(t *testing.T) {
	rand.Seed(1)
	top := NewTree()

	// do a bunch of times
	for round := 0; round < 10; round++ {
		top := NewTree()
		// insert random intervals into the tree
		inserted := 0
		for i := 0; i < 1000; i++ {
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
		countBusy := len(top.BusyIntervals())
		Tassert(t, countBusy == inserted, "should be %v intervals, got %v", inserted, countBusy)
	}

	// loop until all busy intervals are deleted
	busyCount := len(top.BusyIntervals())
	for i := busyCount; i > 0; i-- {
		busyIntervals := top.BusyIntervals()
		Tassert(t, len(busyIntervals) == i, "Expected %d intervals, got %d", i, len(busyIntervals))
		// delete a random interval
		interval := busyIntervals[rand.Intn(len(busyIntervals))]
		top, err := top.Delete(interval)
		Tassert(t, err == nil, err)
		// check that the interval is no longer in the tree
		for _, busyInterval := range top.BusyIntervals() {
			Tassert(t, !busyInterval.Equal(interval), fmt.Sprintf("Expected interval to be deleted, got %v", interval))
		}
		// check that the interval has no conflicts
		conflicts := top.Conflicts(interval, false)
		Tassert(t, len(conflicts) == 0, "Expected 0 conflicts, got %d", len(conflicts))

		// verify the tree
		err = top.Verify(false)
		Tassert(t, err == nil, err)
	}

	// check that all busy intervals are deleted
	busyIntervals := top.BusyIntervals()
	Tassert(t, len(busyIntervals) == 0, "Expected 0 intervals, got %d", len(busyIntervals))

	// check that there is one big free interval
	freeIntervals := top.FreeIntervals()
	Tassert(t, len(freeIntervals) == 1, "Expected 1 interval, got %d", len(freeIntervals))
	start, err := time.Parse(time.RFC3339, TreeStartStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, TreeEndStr)
	Ck(err)
	Tassert(t, freeIntervals[0].Start().Equal(start), fmt.Sprintf("Expected %v, got %v", start, freeIntervals[0].Start()))
	// we use After here instead of Equal because the end time is not exact
	Tassert(t, freeIntervals[0].End().After(end), fmt.Sprintf("Expected %v, got %v", end, freeIntervals[0].End()))
	Tassert(t, freeIntervals[0].Priority() == 0, fmt.Sprintf("Expected %v, got %v", 0, freeIntervals[0].Priority()))
}

func TestRemoveRange(t *testing.T) {
	top := NewTree()

	// insert several intervals into the tree
	i0900_0930 := Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)
	Tassert(t, i0900_0930 != nil, "Failed to insert interval")
	i1000_1100 := Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, i1000_1100 != nil, "Failed to insert interval")
	i1130_1200 := Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T17:00:00Z", 2)
	Tassert(t, i1130_1200 != nil, "Failed to insert interval")

	// RemoveRange removes all intervals that start or end within the
	// given time range.  It returns the removed intervals.  It does not
	// return intervals that are marked as free (priority 0) -- it
	// instead adjusts free intervals to fill gaps in the tree.
	start, err := time.Parse(time.RFC3339, "2024-01-01T09:15:00Z")
	Ck(err)
	end, err := time.Parse(time.RFC3339, "2024-01-01T10:30:00Z")
	Ck(err)
	// check conflicts
	conflicts := top.Conflicts(interval.NewInterval(start, end, 0), true)
	Tassert(t, len(conflicts) == 3, "Expected 3 conflicts, got %v", conflicts)
	// check FindLowerPriority function
	duration := end.Sub(start)
	nodes := top.FindLowerPriority(true, start, end, duration, math.MaxFloat64)
	Tassert(t, len(nodes) == 3, "Expected 3 nodes, got %v", nodes)

	// remove the intervals
	top, removed := top.RemoveRange(start, end)
	Tassert(t, len(removed) > 0, "Expected at least 1 interval, got %d", len(removed))
	Tassert(t, removed[0].Equal(i0900_0930), fmt.Sprintf("Expected %v, got %v", i0900_0930, removed[0]))
	Tassert(t, len(removed) == 2, "Expected 2 intervals, got %d", len(removed))
	Tassert(t, removed[1].Equal(i1000_1100), fmt.Sprintf("Expected %v, got %v", i1000_1100, removed[1]))

	// ShowDot(top, false)

	// check that the 11:30 interval is still in the tree
	intervals := top.BusyIntervals()
	Tassert(t, len(intervals) == 1, "Expected 1 interval, got %d", len(intervals))
	Tassert(t, intervals[0].Equal(i1130_1200), fmt.Sprintf("Expected %v, got %v", i1130_1200, intervals[0]))

	Dump(top, "")

	// check that the free intervals are correct
	freeIntervals := top.FreeIntervals()
	Tassert(t, len(freeIntervals) > 0, "Expected at least 1 free interval, got %d", len(freeIntervals))
	freeExpect := interval.NewInterval(TreeStart, i1130_1200.Start(), 0)
	Tassert(t, freeIntervals[0].Equal(freeExpect), fmt.Sprintf("Expected %v, got %v", freeExpect, freeIntervals[0]))
	Tassert(t, len(freeIntervals) == 2, "Expected 2 free intervals, got %d", len(freeIntervals))
	freeExpect = interval.NewInterval(i1130_1200.End(), TreeEnd, 0)
	Tassert(t, freeIntervals[1].Equal(freeExpect), fmt.Sprintf("Expected %v, got %v", freeExpect, freeIntervals[1]))

	// check that the total number of intervals is correct
	intervals = top.AllIntervals()
	Tassert(t, len(intervals) == 3, "Expected 3 intervals, got %d", len(intervals))

}

func TestShuffle(t *testing.T) {

	top := NewTree()

	// Shuffle inserts a new interval into the tree.  It finds one or
	// more lower-priority intervals using FindLowerPriority, removes
	// and returns them, adjusts the start and end times of the new
	// interval to fit within the free time, and inserts the new
	// interval into the tree.  Shuffle returns the new interval and
	// the removed intervals if successful, otherwise it returns nil,
	// nil. The 'first' parameter determines whether to start
	// searching for free time at the beginning or end of the given
	// time range.  Shuffle does not return intervals that are
	// marked as free (priority 0) -- it instead adjusts free intervals
	// to fill gaps in the tree.

	// insert several intervals into the tree
	i0900_0930 := Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)
	Tassert(t, i0900_0930 != nil, "Failed to insert interval")
	i1000_1100 := Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, i1000_1100 != nil, "Failed to insert interval")
	i1130_1200 := Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T17:00:00Z", 2)
	Tassert(t, i1130_1200 != nil, "Failed to insert interval")

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	Ck(err)

	Dump(top, "")

	// Shuffle a 60 minute interval with priority 3 into the tree near
	// the start time.  Because priority 3 is higher than the priority
	// of the busy interval at 9:00, Shuffle should return the
	// priority 2 interval from 9:00 to 9:30.  The new interval should
	// be inserted into the tree.
	start, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	end, err := time.Parse(time.RFC3339, "2024-01-01T09:30:00Z")
	Ck(err)
	newInterval := interval.NewInterval(start, end, 3)
	newInterval, removed, err := top.Shuffle(true, searchStart, searchEnd, newInterval)
	Tassert(t, err == nil, err)
	Tassert(t, len(removed) == 1, "Expected 1 interval, got %d", len(removed))
	err = Match(removed[0], "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)
	Tassert(t, err == nil, err)

	// XXX ensure that any removed intervals get-re-added before
	// Shuffle exits in case of failure

}
