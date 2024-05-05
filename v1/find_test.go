package v1_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
	"github.com/stevegt/timectl/tree"
	"github.com/stevegt/timectl/util"
)

func TestFindFree(t *testing.T) {
	top := tree.NewTree()

	// insert an interval into the tree -- this should become the left
	// child of the right child of the root node
	tree.Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	tree.Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	tree.Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	goadapt.Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	goadapt.Ck(err)

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
	freeInterval := top.FindFree(true, searchStart, searchEnd, 30*time.Minute)
	goadapt.Tassert(t, freeInterval != nil, "Expected non-nil free interval")
	expectStart, err := time.Parse(time.RFC3339, "2024-01-01T09:30:00Z")
	goadapt.Ck(err)
	expectEnd, err := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	goadapt.Ck(err)
	expectInterval := interval.NewInterval(expectStart, expectEnd, 0)
	goadapt.Tassert(t, freeInterval.Equal(expectInterval), fmt.Sprintf("Expected %s, got %s", expectInterval, freeInterval))

	// find the last free interval that is at least 30 minutes long
	freeInterval = top.FindFree(false, searchStart, searchEnd, 30*time.Minute)
	goadapt.Tassert(t, freeInterval != nil, "Expected non-nil free interval")
	expectStart, err = time.Parse(time.RFC3339, "2024-01-01T17:00:00Z")
	goadapt.Ck(err)
	expectEnd, err = time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	goadapt.Ck(err)
	expectInterval = interval.NewInterval(expectStart, expectEnd, 0)
	goadapt.Tassert(t, freeInterval.Equal(expectInterval), fmt.Sprintf("Expected %s, got %s", expectInterval, freeInterval))

	tree.Verify(t, top, false, false)
}

func TestFindFreeMany(t *testing.T) {
	// This test creates a tree with a number of random intervals and then
	// finds free intervals of varying durations.
	rand.Seed(1)
	top := tree.NewTree()

	// insert several random intervals
	for i := 0; i < 10; i++ {
		start := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
		end := start.Add(time.Duration(rand.Intn(60)) * time.Minute)
		// ignore return value
		tree.Insert(top, start.Format("2006-01-02T15:04:05Z"), end.Format("2006-01-02T15:04:05Z"), 1)
	}

	// Dump(tree, "")

	// find a large number of free intervals of varying durations
	for i := 0; i < 1000; i++ {
		minStart := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
		maxEnd := minStart.Add(time.Duration(rand.Intn(1440)) * time.Minute)
		duration := time.Duration(rand.Intn(60)+1) * time.Minute
		first := rand.Intn(2) == 0
		// t.Logf("minStart: %v, maxEnd: %v, duration: %v, first: %v", minStart, maxEnd, duration, first)
		freeInterval := top.FindFree(first, minStart, maxEnd, duration)
		if freeInterval == nil {
			// sanity check -- try a bunch of times to see if we can find a free interval
			for j := 0; j < 100; j++ {
				start := util.MaxTime(minStart, time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC))
				end := util.MinTime(maxEnd, start.Add(duration))
				if end.Sub(start) < duration {
					continue
				}
				ckInterval := interval.NewInterval(start, end, 1)
				// t.Logf("Trying to find free interval: %v\n", ckInterval)
				if top.Conflicts(ckInterval, false) == nil {
					t.Logf("Found free interval: %v", ckInterval)
					t.Logf("first: %v, minStart: %v, maxEnd: %v, duration: %v", first, minStart, maxEnd, duration)
					for _, iv := range top.AllIntervals() {
						t.Logf("%v", iv)
					}
					t.Fatalf("Expected conflict, got nil")
				}
			}
			continue
		}

		if freeInterval.Duration() < duration {
			t.Fatalf("Expected duration of at least %v, got %v", duration, freeInterval.Duration())
		}

		conflicts := top.Conflicts(freeInterval, false)
		if conflicts != nil {
			t.Logf("Free interval conflict: %v", freeInterval)
			t.Logf("first: %v, minStart: %v, maxEnd: %v, duration: %v", first, minStart, maxEnd, duration)
			for _, iv := range conflicts {
				t.Logf("%v", iv)
			}
			tree.Dump(top, "")
			t.Fatalf("Expected free interval, got conflict")
		}

	}

	tree.Verify(t, top, false, false)

}

// test finding exact interval
func TestFindExact(t *testing.T) {
	top := tree.NewTree()

	// insert an interval into the tree
	iv := tree.NewInterval("2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	ok := top.Insert(iv)
	goadapt.Tassert(t, ok, "Failed to insert interval")

	// showDot(tree, false)

	// FindExact returns the tree node containing the exact interval that
	// matches the given interval, along with the path of ancestor nodes.
	// If the exact interval is not found, then the found node is nil and
	// the path node ends with the node where the interval would be
	// inserted.  If the exact interval is in the root node, then the path
	// is nil.  If the tree is empty, then both are nil.

	path, found := top.FindExact(iv)
	goadapt.Tassert(t, found != nil, "Expected non-nil interval")
	goadapt.Tassert(t, found.GetInterval().Equal(iv), fmt.Sprintf("Expected %v, got %v", iv, found.GetInterval()))
	goadapt.Tassert(t, len(path) == 0, "Expected empty path")

	// try finding an interval that is not in the tree
	iv = tree.NewInterval("2024-01-01T11:30:00Z", "2024-01-01T12:30:00Z", 1)
	path, found = top.FindExact(iv)
	goadapt.Tassert(t, found == nil, "Expected nil interval")
	goadapt.Tassert(t, len(path) == 0, "Expected empty path")

	tree.Verify(t, top, false, false)

}

// FindLowerPriority returns a contiguous set of nodes that have a
// lower priority than the given priority.  The start time of the
// first node is on or before minStart, and the end time of the last
// node is on or after maxEnd.  The nodes must total at least the
// given duration, and may be longer.  If first is true, then the
// search starts at minStart and proceeds in order, otherwise the
// search starts at maxEnd and proceeds in reverse order.
func TestFindLowerPriority(t *testing.T) {
	top := tree.NewTree()

	// insert several intervals into the tree
	i0900_0930 := tree.Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)
	goadapt.Tassert(t, i0900_0930 != nil, "Failed to insert interval")
	i1000_1100 := tree.Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	goadapt.Tassert(t, i1000_1100 != nil, "Failed to insert interval")
	i1130_1200 := tree.Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T17:00:00Z", 2)
	goadapt.Tassert(t, i1130_1200 != nil, "Failed to insert interval")

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	goadapt.Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:00:00Z")
	goadapt.Ck(err)

	// showDot(tree, true)

	// find nodes spanning at least a 60 minute duration and lower
	// than priority 3 near the start time.  because priority 3 is
	// higher than the priority of the busy interval at 9:00,
	// FindLowerPriority should return the priority 2 interval from
	// 9:00 to 9:30 followed by the free interval from 9:30 to 10:00.
	nodes := top.FindLowerPriority(true, searchStart, searchEnd, 60*time.Minute, 3)
	goadapt.Tassert(t, len(nodes) > 0, "Expected at least 1 interval, got %d", len(nodes))
	err = tree.Match(nodes[0].GetInterval(), "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)
	goadapt.Tassert(t, err == nil, err)
	goadapt.Tassert(t, len(nodes) == 2, "Expected 2 nodes, got %d", len(nodes))
	err = tree.Match(nodes[1].GetInterval(), "2024-01-01T9:30:00Z", "2024-01-01T10:00:00Z", 0)
	goadapt.Tassert(t, err == nil, err)

	// find nodes spanning at least a 60 minute duration and lower
	// than priority 2 near the start time.  because priority 2 is not
	// higher than the priority of the busy interval at 9:00,
	// FindLowerPriority should return the priority 0 interval from
	// 9:30 to 10:00 followed by the priority 1 interval from 10:00 to
	// 11:00.
	nodes = top.FindLowerPriority(true, searchStart, searchEnd, 60*time.Minute, 2)
	goadapt.Tassert(t, len(nodes) > 0, "Expected at least 1 interval, got %d", len(nodes))
	err = tree.Match(nodes[0].GetInterval(), "2024-01-01T09:30:00Z", "2024-01-01T10:00:00Z", 0)
	goadapt.Tassert(t, err == nil, err)
	goadapt.Tassert(t, len(nodes) == 2, "Expected 2 nodes, got %d", len(nodes))
	err = tree.Match(nodes[1].GetInterval(), "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	goadapt.Tassert(t, err == nil, err)
	goadapt.Tassert(t, nodes[1].GetInterval() == i1000_1100, "Expected %v, got %v", i1000_1100, nodes[1])

	// find nodes spanning at least a 60 minute duration and lower
	// than priority 2 near the end time.  because priority 2 is not
	// higher than the priority of the interval at 11:30,
	// FindLowerPriority should return the priority 1 interval from
	// 10:00 to 11:00 followed by the priority 0 interval from 11:00
	// to 11:30
	nodes = top.FindLowerPriority(false, searchStart, searchEnd, 60*time.Minute, 2)
	goadapt.Tassert(t, len(nodes) > 0, "Expected at least 1 interval, got %d", len(nodes))
	err = tree.Match(nodes[0].GetInterval(), "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	goadapt.Tassert(t, err == nil, err)
	goadapt.Tassert(t, nodes[0].GetInterval() == i1000_1100, "Expected %v, got %v", i1000_1100, nodes[0])
	goadapt.Tassert(t, len(nodes) == 2, "Expected 2 nodes, got %d", len(nodes))
	err = tree.Match(nodes[1].GetInterval(), "2024-01-01T11:00:00Z", "2024-01-01T11:30:00Z", 0)
	goadapt.Tassert(t, err == nil, err)

	tree.Verify(t, top, false, false)

}
