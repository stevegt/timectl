package find

import (
	"fmt"
	"github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
	"github.com/stevegt/timectl/tree"
	"testing"
	"time"
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
