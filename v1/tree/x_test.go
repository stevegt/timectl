package tree

import (
	"fmt"
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// . "github.com/stevegt/goadapt"

// test merging free nodes
func XXXTestMergeFree(t *testing.T) {
	top := NewTree()

	// split the root node into two free children
	splitAt1200, err := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
	Ck(err)
	top.Interval = nil
	top.Left = &Tree{Interval: interval.NewInterval(TreeStart, splitAt1200, 0).(*interval.IntervalBase)}
	top.Right = &Tree{Interval: interval.NewInterval(splitAt1200, TreeEnd, 0).(*interval.IntervalBase)}

	err = top.Verify()
	Tassert(t, err != nil, "Expected error, got nil")

	// merge the free nodes
	top.mergeFree()

	Verify(t, top)

	// check that the tree has one free interval
	freeIntervals := top.FreeIntervals()
	Tassert(t, len(freeIntervals) == 1, "Expected 1 interval, got %d", len(freeIntervals))
	iv := freeIntervals[0]
	Tassert(t, iv.Start().Equal(TreeStart), fmt.Sprintf("Expected %v, got %v", TreeStart, iv.Start()))
	Tassert(t, iv.End().Equal(TreeEnd), fmt.Sprintf("Expected %v, got %v", TreeEnd, iv.End()))

}

/* deletion algorithm:

- Find the exact interval in the tree
- If the exact interval is not found, then return false
- If the exact interval is found in a leaf node, then remove the node and clear the link in the parent
- if either of the parent's children is a leaf node, then promote the child
- starting with the node's grandparent, walk the tree, merging free nodes as necessary

*/

// test delete
func XXXTestDeleteSimple(t *testing.T) {
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
	err := top.free(found)
	Tassert(t, err == nil, err)

	// check that the interval is no longer in the tree
	intervals := top.BusyIntervals()
	Tassert(t, len(intervals) == 0, "Expected 0 intervals, got %d", len(intervals))
	// we haven't merged free nodes yet, so there should be three free nodes
	freeIntervals := top.FreeIntervals()
	Tassert(t, len(freeIntervals) == 3, "Expected 3 free intervals, got %d", len(freeIntervals))

	// merge the free nodes
	top.mergeFree()
	// there should now be one free interval
	freeIntervals = top.FreeIntervals()
	Tassert(t, len(freeIntervals) == 1, "Expected 1 interval, got %d", len(freeIntervals))
	freeInterval := freeIntervals[0]
	Tassert(t, freeInterval.Start().Equal(TreeStart), fmt.Sprintf("Expected %v, got %v", TreeStart, freeInterval.Start()))
	Tassert(t, freeInterval.End().Equal(TreeEnd), fmt.Sprintf("Expected %v, got %v", TreeEnd, freeInterval.End()))

	Verify(t, top)

	/*
		// delete is simply a process of finding and freeing the target
		// interval and then merging free nodes.
		ok = tree.Insert(interval)
		Tassert(t, ok, "Failed to insert interval")
		deletedInterval = tree.delete(interval)
		Tassert(t, deletedInterval != nil, "Expected non-nil interval")
		Tassert(t, deletedInterval.Equal(interval), fmt.Sprintf("Expected %v, got %v", interval, deletedInterval))
		allIntervals := tree.AllIntervals()
		Tassert(t, len(allIntervals) == 1, "Expected 1 interval, got %d", len(allIntervals))
		Tassert(t, allIntervals[0].Equal(freeInterval), fmt.Sprintf("Expected %v, got %v", freeInterval, allIntervals[0]))

		verify(t, tree)
	*/

}

/*
// test complex delete
func TestDeleteComplex(t *testing.T) {
	rand.Seed(1)
	top := NewTree()

	// insert several random intervals
	for i := 0; i < 10; i++ {
		start := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
		end := start.Add(time.Duration(rand.Intn(60)) * time.Minute)
		// ignore return value
		insert(tree, start.Format("2006-01-02T15:04:05Z"), end.Format("2006-01-02T15:04:05Z"), 1)
	}

	// loop until all busy intervals are deleted
	busyCount := len(tree.BusyIntervals())
	for i := busyCount; i > 0; i-- {
		busyIntervals := tree.BusyIntervals()
		Tassert(t, len(busyIntervals) == i, "Expected %d intervals, got %d", i, len(busyIntervals))
		// delete a random interval
		interval := busyIntervals[rand.Intn(len(busyIntervals))]
		ok := tree.Delete(interval)
		Tassert(t, ok, "Failed to delete interval")
		// check that the interval is no longer in the tree
		for _, busyInterval := range tree.BusyIntervals() {
			Tassert(t, !busyInterval.Equal(interval), fmt.Sprintf("Expected interval to be deleted, got %v", interval))
		}
		// check that the interval has no conflicts
		conflicts := tree.Conflicts(interval, false)
		Tassert(t, len(conflicts) == 0, "Expected 0 conflicts, got %d", len(conflicts))
		// check that there are no adjacent free intervals
		freeIntervals := tree.FreeIntervals()
		prev := freeIntervals[0]
		for j := 1; j < len(freeIntervals); j++ {
			if prev.End().Equal(freeIntervals[j].Start()) {
				t.Logf("prev: %v", prev)
				t.Logf("next: %v", freeIntervals[j])
				t.Fatalf("Expected no adjacent free intervals")
			}
			prev = freeIntervals[j]
		}
		// check that there are no gaps between intervals
		allIntervals := tree.AllIntervals()
		prev = allIntervals[0]
		for j := 1; j < len(allIntervals); j++ {
			if prev.End().Before(allIntervals[j].Start()) {
				t.Logf("prev: %v", prev)
				t.Logf("next: %v", allIntervals[j])
				t.Fatalf("Expected no gaps between intervals")
			}
			prev = allIntervals[j]
		}
	}

	// check that all busy intervals are deleted
	busyIntervals := tree.BusyIntervals()
	Tassert(t, len(busyIntervals) == 0, "Expected 0 intervals, got %d", len(busyIntervals))

	// check that there is one big free interval
	freeIntervals := tree.FreeIntervals()
	Tassert(t, len(freeIntervals) == 1, "Expected 1 interval, got %d", len(freeIntervals))
	start, err := time.Parse(time.RFC3339, TreeStartStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, TreeEndStr)
	Ck(err)
	Tassert(t, freeIntervals[0].Start().Equal(start), fmt.Sprintf("Expected %v, got %v", start, freeIntervals[0].Start()))
	// XXX we use After here instead of Equal because the end time is not exact
	Tassert(t, freeIntervals[0].End().After(end), fmt.Sprintf("Expected %v, got %v", end, freeIntervals[0].End()))
	Tassert(t, freeIntervals[0].Priority() == 0, fmt.Sprintf("Expected %v, got %v", 0, freeIntervals[0].Priority()))
}

func TestRemoveRange(t *testing.T) {
	top := NewTree()

	// insert several intervals into the tree
	i0900_0930 := insert(tree, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)
	Tassert(t, i0900_0930 != nil, "Failed to insert interval")
	i1000_1100 := insert(tree, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, i1000_1100 != nil, "Failed to insert interval")
	i1130_1200 := insert(tree, "2024-01-01T11:30:00Z", "2024-01-01T17:00:00Z", 2)
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
	conflicts := tree.Conflicts(interval.NewInterval(start, end, 0), true)
	Tassert(t, len(conflicts) == 3, "Expected 3 conflicts, got %v", conflicts)
	// remove the intervals
	removed := tree.RemoveRange(start, end)
	Tassert(t, len(removed) > 0, "Expected at least 1 interval, got %d", len(removed))
	Tassert(t, removed[0].Equal(i0900_0930), fmt.Sprintf("Expected %v, got %v", i0900_0930, removed[0]))
	Tassert(t, len(removed) == 2, "Expected 2 intervals, got %d", len(removed))
	Tassert(t, removed[1].Equal(i1000_1100), fmt.Sprintf("Expected %v, got %v", i1000_1100, removed[1]))

	// check that the 11:30 interval is still in the tree
	intervals := tree.BusyIntervals()
	Tassert(t, len(intervals) == 1, "Expected 1 interval, got %d", len(intervals))
	Tassert(t, intervals[0].Equal(i1130_1200), fmt.Sprintf("Expected %v, got %v", i1130_1200, intervals[0]))

	Dump(tree, "")

	// check that the free intervals are correct
	freeIntervals := tree.FreeIntervals()
	Tassert(t, len(freeIntervals) > 0, "Expected at least 1 free interval, got %d", len(freeIntervals))
	freeExpect := interval.NewInterval(TreeStart, i1130_1200.Start(), 0)
	Tassert(t, freeIntervals[0].Equal(freeExpect), fmt.Sprintf("Expected %v, got %v", freeExpect, freeIntervals[0]))
	Tassert(t, len(freeIntervals) == 2, "Expected 2 free intervals, got %d", len(freeIntervals))
	freeExpect = interval.NewInterval(i1130_1200.End(), TreeEnd, 0)
	Tassert(t, freeIntervals[1].Equal(freeExpect), fmt.Sprintf("Expected %v, got %v", freeExpect, freeIntervals[1]))

	// check that the total number of intervals is correct
	intervals = tree.AllIntervals()
	Tassert(t, len(intervals) == 3, "Expected 3 intervals, got %d", len(intervals))

}

func XXXTestShuffle(t *testing.T) {

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
	i0900_0930 := insert(tree, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)
	Tassert(t, i0900_0930 != nil, "Failed to insert interval")
	i1000_1100 := insert(tree, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, i1000_1100 != nil, "Failed to insert interval")
	i1130_1200 := insert(tree, "2024-01-01T11:30:00Z", "2024-01-01T17:00:00Z", 2)
	Tassert(t, i1130_1200 != nil, "Failed to insert interval")

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	Ck(err)

	Dump(tree, "")

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
	newInterval, removed, err := tree.Shuffle(true, searchStart, searchEnd, newInterval)
	Tassert(t, err == nil, err)
	Tassert(t, len(removed) == 1, "Expected 1 interval, got %d", len(removed))
	err = match(removed[0], "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)

	// XXX ensure that any removed intervals get-re-added before
	// Shuffle exits in case of failure

}
*/
