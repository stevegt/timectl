package concurrent

import (
	"fmt"
	"github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
	tree2 "github.com/stevegt/timectl/tree"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestConcurrent(t *testing.T) {
	// This test creates a tree with a number of random intervals and then
	// finds free intervals of varying durations.  It does this in
	// multiple goroutines in order to test thread safety.
	rand.Seed(1)
	tree := tree2.NewTree()

	// insert several random intervals in multiple goroutines
	insertMap := sync.Map{}
	wgInsert := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wgInsert.Add(1)
		go func(i int) {
			// wait a random amount of time
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			// try to insert a random interval
			start := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
			end := start.Add(time.Duration(rand.Intn(59)+1) * time.Minute)
			iv := tree2.Insert(tree, start.Format("2006-01-02T15:04:05Z"), end.Format("2006-01-02T15:04:05Z"), 1)
			if iv != nil {
				insertMap.Store(i, iv)
			}
			wgInsert.Done()
		}(i)
	}

	// find free intervals while the tree is being
	// modified by the insert goroutines
	foundCount := 0

	for i := 0; i < 1000; i++ {
		minStart := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
		maxEnd := minStart.Add(time.Duration(rand.Intn(1440)) * time.Minute)
		duration := time.Duration(rand.Intn(60)+1) * time.Minute
		first := rand.Intn(2) == 0
		freeInterval := tree.FindFree(first, minStart, maxEnd, duration)
		if freeInterval != nil {
			foundCount++
		}
	}

	goadapt.Tassert(t, foundCount > 0, "Expected at least one free interval")

	// wait for all insert goroutines to finish
	wgInsert.Wait()

	// copy the intervals from insertMap to a slice
	var inserted []interval.Interval
	insertMap.Range(func(key, value any) bool {
		inserted = append(inserted, value.(interval.Interval))
		return true
	})

	size := len(inserted)
	goadapt.Tassert(t, size > 0, "Expected at least one interval")
	goadapt.Pf("Inserted %d intervals\n", size)

	// check that all intervals were inserted
	busyLen := len(tree.BusyIntervals())
	goadapt.Tassert(t, busyLen == size, "Expected %d intervals, got %d", size, busyLen)

	for _, expect := range inserted {
		// we expect 1 conflict for each interval
		conflicts := tree.Conflicts(expect, false)
		goadapt.Tassert(t, len(conflicts) == 1, "Expected 1 conflict, got %d", len(conflicts))
		// check that the conflict is the expected interval
		goadapt.Tassert(t, conflicts[0].Equal(expect), fmt.Sprintf("Expected %v, got %v", expect, conflicts[0]))
	}

	tree2.Verify(t, tree, false, false)

}
