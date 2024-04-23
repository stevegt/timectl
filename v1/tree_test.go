package timectl

import (
	"fmt"
	"math/rand"
	"sync"
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

// get is a test helper function that returns the tree node at the
// given path in the tree.  pathStr is the path to the node in the
// tree, where 'l' means to go left and 'r' means to go right.  An
// empty pathStr means to return the root node.
func get(tree *Tree, pathStr string) *Tree {
	path := []rune(pathStr)
	if len(path) == 0 {
		return tree
	}
	switch path[0] {
	case 'l':
		return get(tree.left, string(path[1:]))
	case 'r':
		return get(tree.right, string(path[1:]))
	default:
		Assert(false, "Invalid path %v", pathStr)
	}
	return nil
}

// expect is a test helper function that checks if the given tree
// node's interval has the expected start and end times and priority.
// pathStr is the path to the node in the tree, where 'l'
// means to go left and 'r' means to go right.  An empty pathStr means
// to check the root node.
func expect(tree *Tree, pathStr, startStr, endStr string, priority float64) error {
	node := get(tree, pathStr)
	if node == nil {
		return fmt.Errorf("no node at path: %v", pathStr)
	}
	nodeInterval := node.Interval()
	if nodeInterval.Priority() != priority {
		return fmt.Errorf("Expected priority=%v, got priority=%v", priority, nodeInterval.Priority())
	}
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	ev := NewInterval(start, end, priority)
	if !node.Interval().Equal(ev) {
		return fmt.Errorf("Expected %v, got %v", ev, node.Interval())
	}
	return nil
}

// insertExpect is a test helper function that inserts an interval
// into the tree and checks if the tree has the expected structure.
func insertExpect(tree *Tree, pathStr, startStr, endStr string, priority float64) error {
	interval := insert(tree, startStr, endStr, priority)
	if interval == nil {
		return fmt.Errorf("Failed to insert interval")
	}
	return expect(tree, pathStr, startStr, endStr, priority)
}

// newInterval is a test helper function that creates a new interval
// with the given start and end times and priority content.
func newInterval(startStr, endStr string, priority float64) Interval {
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	return NewInterval(start, end, priority)
}

// insert is a test helper function that inserts an interval into the
// tree and returns the interval that was inserted.
func insert(tree *Tree, startStr, endStr string, priority float64) Interval {
	interval := newInterval(startStr, endStr, priority)
	// Insert adds a new interval to the tree, adjusting the structure as
	// necessary.  Insertion fails if the new interval conflicts with any
	// existing interval in the tree.
	ok := tree.Insert(interval)
	if !ok {
		return nil
	}
	return interval
}

// match is a test helper function that checks if the given interval
// has the expected start and end times and priority.
func match(interval Interval, startStr, endStr string, priority float64) error {
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	ev := NewInterval(start, end, priority)
	if !interval.Equal(ev) {
		return fmt.Errorf("Expected %v, got %v", ev, interval)
	}
	return nil
}

// Test a tree node with children.
// A tree is a tree of busy and free intervals that
// span the entire range from treeStart to treeEnd.
func TestTreeStructure(t *testing.T) {

	tree := NewTree()
	// insert interval into empty tree -- this should become the left
	// child of the right child of the root node
	err := insertExpect(tree, "rl", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)
	// the other nodes should be non-busy
	err = expect(tree, "l", TreeStartStr, "2024-01-01T10:00:00Z", 0)
	Tassert(t, err == nil, err)
	err = expect(tree, "rr", "2024-01-01T11:00:00Z", TreeEndStr, 0)

}

// TestInsertConflict tests inserting an interval that conflicts with
// an existing interval in the tree.
func TestInsertConflict(t *testing.T) {

	tree := NewTree()

	// insert an interval into the tree
	err := insertExpect(tree, "rl", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)

	// insert a conflicting interval
	interval := insert(tree, "2024-01-01T10:30:00Z", "2024-01-01T11:30:00Z", 1)
	Tassert(t, interval == nil, "Expected nil interval")

}

func TestConflicts(t *testing.T) {
	tree := NewTree()

	// insert several intervals into the tree
	i1000_1100 := insert(tree, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, i1000_1100 != nil, "Failed to insert interval")
	i1130_1200 := insert(tree, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Tassert(t, i1130_1200 != nil, "Failed to insert interval")
	i0900_0930 := insert(tree, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)
	Tassert(t, i0900_0930 != nil, "Failed to insert interval")

	// create a new interval that overlaps the first interval
	i1030_1130 := newInterval("2024-01-01T10:30:00Z", "2024-01-01T11:30:00Z", 1)
	// get conflicts for the new interval
	conflicts := tree.Conflicts(i1030_1130)
	Tassert(t, len(conflicts) == 1, "Expected 1 conflict, got %d", len(conflicts))
	Tassert(t, conflicts[0].Equal(i1000_1100), fmt.Sprintf("Expected %v, got %v", i1000_1100, conflicts[0]))

	// ensure BusyIntervals() returns all intervals
	intervals := tree.BusyIntervals()
	Tassert(t, len(intervals) == 3, "Expected 3 intervals, got %d", len(intervals))
	Tassert(t, intervals[0].Equal(i0900_0930), fmt.Sprintf("Expected %v, got %v", i0900_0930, intervals[0]))
	Tassert(t, intervals[1].Equal(i1000_1100), fmt.Sprintf("Expected %v, got %v", i1000_1100, intervals[1]))
	Tassert(t, intervals[2].Equal(i1130_1200), fmt.Sprintf("Expected %v, got %v", i1130_1200, intervals[2]))

}

func TestFindFree(t *testing.T) {
	tree := NewTree()

	// insert an interval into the tree -- this should become the left
	// child of the right child of the root node
	err := insertExpect(tree, "rl", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)
	err = insertExpect(tree, "rrrl", "2024-01-01T12:00:00Z", "2024-01-01T13:00:00Z", 1)
	Tassert(t, err == nil, err)
	err = insertExpect(tree, "lrl", "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)
	Tassert(t, err == nil, err)

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	Ck(err)

	// FindFree returns an interval that has the given duration.  The interval
	// starts as early as possible if first is true, and as late as possible
	// if first is false.  The minStart and maxEnd times are inclusive.
	// The duration is exclusive.
	//
	// This function works by walking the tree in a depth-first manner,
	// following the left child first if first is set, otherwise following
	// the right child first.

	// dump(tree, "")
	// find the first free interval that is at least 30 minutes long
	freeInterval := tree.FindFree(true, searchStart, searchEnd, 30*time.Minute)
	Tassert(t, freeInterval != nil, "Expected non-nil free interval")
	expectStart, err := time.Parse(time.RFC3339, "2024-01-01T09:30:00Z")
	Ck(err)
	expectEnd, err := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	Ck(err)
	expectInterval := NewInterval(expectStart, expectEnd, 0)
	Tassert(t, freeInterval.Equal(expectInterval), fmt.Sprintf("Expected %s, got %s", expectInterval, freeInterval))

	// find the last free interval that is at least 30 minutes long
	freeInterval = tree.FindFree(false, searchStart, searchEnd, 30*time.Minute)
	Tassert(t, freeInterval != nil, "Expected non-nil free interval")
	expectStart, err = time.Parse(time.RFC3339, "2024-01-01T17:00:00Z")
	Ck(err)
	expectEnd, err = time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	Ck(err)
	expectInterval = NewInterval(expectStart, expectEnd, 0)
	Tassert(t, freeInterval.Equal(expectInterval), fmt.Sprintf("Expected %s, got %s", expectInterval, freeInterval))

}

func TestFindFreeMany(t *testing.T) {
	// This test creates a tree with a number of random intervals and then
	// finds free intervals of varying durations.
	rand.Seed(1)
	tree := NewTree()

	// insert several random intervals
	for i := 0; i < 10; i++ {
		start := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
		end := start.Add(time.Duration(rand.Intn(60)) * time.Minute)
		// ignore return value
		insert(tree, start.Format("2006-01-02T15:04:05Z"), end.Format("2006-01-02T15:04:05Z"), 1)
	}

	// dump(tree, "")

	// find a large number of free intervals of varying durations
	for i := 0; i < 1000; i++ {
		minStart := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
		maxEnd := minStart.Add(time.Duration(rand.Intn(1440)) * time.Minute)
		duration := time.Duration(rand.Intn(60)+1) * time.Minute
		first := rand.Intn(2) == 0
		// t.Logf("minStart: %v, maxEnd: %v, duration: %v, first: %v", minStart, maxEnd, duration, first)
		freeInterval := tree.FindFree(first, minStart, maxEnd, duration)
		if freeInterval == nil {
			// sanity check -- try a bunch of times to see if we can find a free interval
			for j := 0; j < 100; j++ {
				start := MaxTime(minStart, time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC))
				end := MinTime(maxEnd, start.Add(duration))
				if end.Sub(start) < duration {
					continue
				}
				ckInterval := NewInterval(start, end, 1)
				// t.Logf("Trying to find free interval: %v\n", ckInterval)
				if tree.Conflicts(ckInterval) == nil {
					t.Logf("Found free interval: %v", ckInterval)
					t.Logf("first: %v, minStart: %v, maxEnd: %v, duration: %v", first, minStart, maxEnd, duration)
					for _, interval := range tree.AllIntervals() {
						t.Logf("%v", interval)
					}
					t.Fatalf("Expected conflict, got nil")
				}
			}
			continue
		}

		if freeInterval.Duration() < duration {
			t.Fatalf("Expected duration of at least %v, got %v", duration, freeInterval.Duration())
		}

		conflicts := tree.Conflicts(freeInterval)
		if conflicts != nil {
			t.Logf("Free interval conflict: %v", freeInterval)
			t.Logf("first: %v, minStart: %v, maxEnd: %v, duration: %v", first, minStart, maxEnd, duration)
			for _, interval := range conflicts {
				t.Logf("%v", interval)
			}
			dump(tree, "")
			t.Fatalf("Expected free interval, got conflict")
		}

	}
}

func TestConcurrent(t *testing.T) {
	// This test creates a tree with a number of random intervals and then
	// finds free intervals of varying durations.  It does this in
	// multiple goroutines in order to test thread safety.
	rand.Seed(1)
	tree := NewTree()

	size := 10

	// insert several random intervals in multiple goroutines
	insertMap := sync.Map{}
	wgInsert := sync.WaitGroup{}
	for i := 0; i < size; i++ {
		wgInsert.Add(1)
		go func(i int) {
			// retry until we can insert an interval
			for {
				// wait a random amount of time
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				start := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
				end := start.Add(time.Duration(rand.Intn(60)) * time.Minute)
				interval := insert(tree, start.Format("2006-01-02T15:04:05Z"), end.Format("2006-01-02T15:04:05Z"), 1)
				if interval == nil {
					continue
				}
				insertMap.Store(i, interval)
				break
			}
			wgInsert.Done()
		}(i)
	}

	// find free intervals in multiple goroutines while the tree is being
	// modified by the insert goroutines
	foundCount := 0

	for i := 0; i < 1000; i++ {
		minStart := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
		maxEnd := minStart.Add(time.Duration(rand.Intn(1440)) * time.Minute)
		duration := time.Duration(rand.Intn(60)+1) * time.Minute
		first := rand.Intn(2) == 0
		// ignore return value
		freeInterval := tree.FindFree(first, minStart, maxEnd, duration)
		if freeInterval != nil {
			foundCount++
		}
	}

	Tassert(t, foundCount > 0, "Expected at least one free interval")

	// wait for all insert goroutines to finish
	wgInsert.Wait()

	// copy the intervals from insertMap to a slice
	inserted := make([]*IntervalBase, size)
	insertMap.Range(func(key, value any) bool {
		inserted[key.(int)] = value.(*IntervalBase)
		return true
	})

	// check that all intervals were inserted
	Tassert(t, len(inserted) == size, "Expected %d intervals, got %d", size, len(inserted))
	busyLen := len(tree.BusyIntervals())
	Tassert(t, busyLen == size, "Expected %d intervals, got %d", size, busyLen)

	for _, expect := range inserted {
		// we expect 1 conflict for each interval
		conflicts := tree.Conflicts(expect)
		Tassert(t, len(conflicts) == 1, "Expected 1 conflict, got %d", len(conflicts))
		// check that the conflict is the expected interval
		Tassert(t, conflicts[0].Equal(expect), fmt.Sprintf("Expected %v, got %v", expect, conflicts[0]))
	}
}

// ConcreteInterval tests the Interval interface and IntervalBase type.
type ConcreteInterval struct {
	*IntervalBase
}

func NewConcreteInterval(start, end time.Time, priority float64) *ConcreteInterval {
	interval := &ConcreteInterval{
		IntervalBase: NewInterval(start, end, priority).(*IntervalBase),
	}
	return interval
}

func TestInterface(t *testing.T) {
	// This test checks the basic functionality of the Interval interface
	// and IntervalBase type.
	tree := NewTree()

	start, err := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	Ck(err)
	end, err := time.Parse(time.RFC3339, "2024-01-01T11:00:00Z")
	Ck(err)
	interval := NewConcreteInterval(start, end, 1)
	Tassert(t, interval.Start().Equal(start), fmt.Sprintf("Expected %v, got %v", start, interval.Start()))
	Tassert(t, interval.End().Equal(end), fmt.Sprintf("Expected %v, got %v", end, interval.End()))
	Tassert(t, interval.Priority() == 1, fmt.Sprintf("Expected %v, got %v", 1, interval.Priority()))

	// insert the interval into the tree
	ok := tree.Insert(interval)
	Tassert(t, ok, "Failed to insert interval")

	// dump(tree, "")

	// check that the interval is in the tree
	intervals := tree.BusyIntervals()
	Tassert(t, len(intervals) == 1, "Expected 1 interval, got %d", len(intervals))
	Tassert(t, intervals[0].Equal(interval), fmt.Sprintf("Expected %v, got %v", interval, intervals[0]))

	// check that the interval is returned by AllIntervals
	intervals = tree.AllIntervals()
	Tassert(t, len(intervals) == 3, "Expected 3 intervals, got %d", len(intervals))
	Tassert(t, intervals[1].Equal(interval), fmt.Sprintf("Expected %v, got %v", interval, intervals[1]))

}

// test delete
func TestDelete(t *testing.T) {
	tree := NewTree()

	// insert an interval into the tree
	err := insertExpect(tree, "rl", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)

	// delete the interval
	interval := newInterval("2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	ok := tree.Delete(interval)
	Tassert(t, ok, "Failed to delete interval")

	// check that the interval is no longer in the tree
	intervals := tree.BusyIntervals()
	Tassert(t, len(intervals) == 0, "Expected 0 intervals, got %d", len(intervals))
}

// test complex delete
func TestDeleteComplex(t *testing.T) {
	rand.Seed(1)
	tree := NewTree()

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
		conflicts := tree.Conflicts(interval)
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

// FindFreePriority works similarly to FindFree, but it returns a
// contiguous set of intervals that are either free or have a lower
// priority than the given priority.  The intervals are returned in
// order of start time.  The minStart and maxEnd times are inclusive.
//
// FindFreePriority pseudocode:
/*
- Start the search at `minStart` and iterate through the tree in order
  (or start at maxEnd and traverse in reverse order if `first` is false)
- If the current interval has a lower priority than the given
  priority, add it to a candidate list
- If the current interval has a higher priority than the given
  priority, clear the candidate list
- If the candidate list has intervals that are contiguous with the
  current interval, add the current interval to the candidate list,
  else clear the candidate list
- If the candidate list has intervals that span the given duration,
  return the candidate list
*/
func TestFindFreePriority(t *testing.T) {
	tree := NewTree()

	// insert several intervals into the tree
	i1000_1100 := insert(tree, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, i1000_1100 != nil, "Failed to insert interval")
	i1130_1200 := insert(tree, "2024-01-01T11:30:00Z", "2024-01-01T17:00:00Z", 2)
	Tassert(t, i1130_1200 != nil, "Failed to insert interval")
	i0900_0930 := insert(tree, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)
	Tassert(t, i0900_0930 != nil, "Failed to insert interval")

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	Ck(err)

	dump(tree, "")

	// find intervals for a 60 minute duration and priority 3 near the
	// start time.  because priority 3 is higher than the priority of
	// the busy interval at 9:00, FindFreePriority should return the
	// priority 2 interval from 9:00 to 9:30 followed by the free
	// interval from 9:30 to 10:00.
	intervals := tree.FindFreePriority(true, searchStart, searchEnd, 60*time.Minute, 3)
	t.Logf("intervals found that are lower priority than 3:")
	for _, interval := range intervals {
		t.Logf("%v", interval)
	}
	Tassert(t, len(intervals) > 0, "Expected at least 1 interval, got %d", len(intervals))
	err = match(intervals[0], "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)
	Tassert(t, err == nil, err)
	Tassert(t, len(intervals) == 2, "Expected 2 intervals, got %d", len(intervals))
	err = match(intervals[1], "2024-01-01T9:30:00Z", "2024-01-01T10:00:00Z", 0)
	Tassert(t, err == nil, err)

	// find intervals for a 60 minute duration and priority 2 near the
	// start time.  because priority 2 is not higher than the priority of
	// the busy interval at 9:00, FindFreePriority should return the free
	// interval from 9:30 to 10:00 followed by the priority 1 interval
	// from 10:00 to 11:00.
	intervals = tree.FindFreePriority(true, searchStart, searchEnd, 60*time.Minute, 2)
	t.Logf("intervals found that are lower priority than 2:")
	for _, interval := range intervals {
		t.Logf("%v", interval)
	}
	Tassert(t, len(intervals) > 0, "Expected at least 1 interval, got %d", len(intervals))
	err = match(intervals[0], "2024-01-01T09:30:00Z", "2024-01-01T10:00:00Z", 0)
	Tassert(t, err == nil, err)
	Tassert(t, len(intervals) == 2, "Expected 2 intervals, got %d", len(intervals))
	err = match(intervals[1], "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)
	Tassert(t, intervals[1] == i1000_1100, "Expected %v, got %v", i1000_1100, intervals[1])

	// find intervals for a 60 minute duration and priority 2 near the
	// end time.  because priority 2 is not higher than the priority
	// of the interval at 11:30, FindFreePriority should return the
	// priority 1 interval from 10:00 to 11:00 followed by the free
	// interval from 11:00 to 11:30
	intervals = tree.FindFreePriority(false, searchStart, searchEnd, 60*time.Minute, 2)
	t.Logf("intervals found that are lower priority than 2 near end:")
	for _, interval := range intervals {
		t.Logf("%v", interval)
	}
	Tassert(t, len(intervals) > 0, "Expected at least 1 interval, got %d", len(intervals))
	err = match(intervals[0], "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Tassert(t, err == nil, err)
	Tassert(t, intervals[0] == i1000_1100, "Expected %v, got %v", i1000_1100, intervals[1])
	Tassert(t, len(intervals) == 2, "Expected 2 intervals, got %d", len(intervals))
	err = match(intervals[1], "2024-01-01T11:00:00Z", "2024-01-01T11:30:00Z", 0)
	Tassert(t, err == nil, err)

}