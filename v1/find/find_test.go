package find

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

	tree.Verify(t, top)
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

	tree.Verify(t, top)

}
