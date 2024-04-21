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
// node's interval has the expected start and end times and payload
// content. pathStr is the path to the node in the tree, where 'l'
// means to go left and 'r' means to go right.  An empty pathStr means
// to check the root node.
func expect(tree *Tree, pathStr, startStr, endStr string, payload any) error {
	node := get(tree, pathStr)
	if node == nil {
		return fmt.Errorf("no node at path: %v", pathStr)
	}
	nodeInterval := node.Interval()
	if nodeInterval.Payload() != payload {
		return fmt.Errorf("Expected payload=%v, got payload=%v", payload, nodeInterval.Payload())
	}
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	ev := NewInterval(start, end, payload)
	if !node.Interval().Equal(ev) {
		return fmt.Errorf("Expected %v, got %v", ev, node.Interval())
	}
	return nil
}

// insertExpect is a test helper function that inserts an interval
// into the tree and checks if the tree has the expected structure.
func insertExpect(tree *Tree, pathStr, startStr, endStr string, payload any) error {
	interval := insert(tree, startStr, endStr, payload)
	if interval == nil {
		return fmt.Errorf("Failed to insert interval")
	}
	return expect(tree, pathStr, startStr, endStr, payload)
}

// newInterval is a test helper function that creates a new interval
// with the given start and end times and payload content.
func newInterval(startStr, endStr string, payload any) *Interval {
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	return NewInterval(start, end, payload)
}

// insert is a test helper function that inserts an interval into the
// tree and returns the interval that was inserted.
func insert(tree *Tree, startStr, endStr string, payload any) *Interval {
	interval := newInterval(startStr, endStr, payload)
	// Insert adds a new interval to the tree, adjusting the structure as
	// necessary.  Insertion fails if the new interval conflicts with any
	// existing interval in the tree.
	ok := tree.Insert(interval)
	if !ok {
		return nil
	}
	return interval
}

// Test a tree node with children.
// A tree is a tree of busy and free intervals that
// span the entire range from treeStart to treeEnd.
func TestTreeStructure(t *testing.T) {

	tree := NewTree()
	// insert interval into empty tree -- this should become the left
	// child of the right child of the root node
	err := insertExpect(tree, "rl", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	Tassert(t, err == nil, err)
	// the other nodes should be non-busy
	err = expect(tree, "l", TreeStartStr, "2024-01-01T10:00:00Z", nil)
	Tassert(t, err == nil, err)
	err = expect(tree, "rr", "2024-01-01T11:00:00Z", TreeEndStr, nil)

}

// TestInsertConflict tests inserting an interval that conflicts with
// an existing interval in the tree.
func TestInsertConflict(t *testing.T) {

	tree := NewTree()

	// insert an interval into the tree
	err := insertExpect(tree, "rl", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	Tassert(t, err == nil, err)

	// insert a conflicting interval
	interval := insert(tree, "2024-01-01T10:30:00Z", "2024-01-01T11:30:00Z", true)
	Tassert(t, interval == nil, "Expected nil interval")

}

func TestConflicts(t *testing.T) {
	tree := NewTree()

	// insert several intervals into the tree
	i1000_1100 := insert(tree, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	Tassert(t, i1000_1100 != nil, "Failed to insert interval")
	i1130_1200 := insert(tree, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", true)
	Tassert(t, i1130_1200 != nil, "Failed to insert interval")
	i0900_0930 := insert(tree, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", true)
	Tassert(t, i0900_0930 != nil, "Failed to insert interval")

	// create a new interval that overlaps the first interval
	i1030_1130 := newInterval("2024-01-01T10:30:00Z", "2024-01-01T11:30:00Z", true)
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
	err := insertExpect(tree, "rl", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	Tassert(t, err == nil, err)
	err = insertExpect(tree, "rrrl", "2024-01-01T12:00:00Z", "2024-01-01T13:00:00Z", true)
	Tassert(t, err == nil, err)
	err = insertExpect(tree, "lrl", "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", true)
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
	expectInterval := NewInterval(expectStart, expectEnd, nil)
	Tassert(t, freeInterval.Equal(expectInterval), fmt.Sprintf("Expected %s, got %s", expectInterval, freeInterval))

	// find the last free interval that is at least 30 minutes long
	freeInterval = tree.FindFree(false, searchStart, searchEnd, 30*time.Minute)
	Tassert(t, freeInterval != nil, "Expected non-nil free interval")
	expectStart, err = time.Parse(time.RFC3339, "2024-01-01T17:00:00Z")
	Ck(err)
	expectEnd, err = time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	Ck(err)
	expectInterval = NewInterval(expectStart, expectEnd, nil)
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
		insert(tree, start.Format("2006-01-02T15:04:05Z"), end.Format("2006-01-02T15:04:05Z"), true)
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
				start := maxTime(minStart, time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC))
				end := minTime(maxEnd, start.Add(duration))
				if end.Sub(start) < duration {
					continue
				}
				ckInterval := NewInterval(start, end, true)
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
				interval := insert(tree, start.Format("2006-01-02T15:04:05Z"), end.Format("2006-01-02T15:04:05Z"), true)
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
	inserted := make([]*Interval, size)
	insertMap.Range(func(key, value any) bool {
		inserted[key.(int)] = value.(*Interval)
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
