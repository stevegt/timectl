package v1

import (
	"fmt"
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/v2/interval"
	"github.com/stevegt/timectl/v2/tree"
)

// ConcreteInterval tests the interval.Interval interface and interval.IntervalBase type.
type ConcreteInterval struct {
	*interval.IntervalBase
}

func NewConcreteInterval(start, end time.Time, priority float64) *ConcreteInterval {
	iv := &ConcreteInterval{
		IntervalBase: interval.NewInterval(start, end, priority).(*interval.IntervalBase),
	}
	return iv
}

func TestInterface(t *testing.T) {
	// This test checks the basic functionality of the interval.Interval interface
	// and interval.IntervalBase type.
	top := tree.NewTree()

	start, err := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	Ck(err)
	end, err := time.Parse(time.RFC3339, "2024-01-01T11:00:00Z")
	Ck(err)
	iv := NewConcreteInterval(start, end, 1)
	Tassert(t, iv.Start().Equal(start), fmt.Sprintf("Expected %v, got %v", start, iv.Start()))
	Tassert(t, iv.End().Equal(end), fmt.Sprintf("Expected %v, got %v", end, iv.End()))
	Tassert(t, iv.Priority() == 1, fmt.Sprintf("Expected %v, got %v", 1, iv.Priority()))

	// insert the interval into the tree
	_, err = top.Insert(iv)
	Tassert(t, err == nil, fmt.Sprintf("Expected nil, got %v", err))

	// Dump(tree, "")

	// check that the interval is in the tree
	intervals := top.BusyIntervals()
	Tassert(t, len(intervals) == 1, "Expected 1 interval, got %d", len(intervals))
	Tassert(t, intervals[0].Equal(iv), fmt.Sprintf("Expected %v, got %v", iv, intervals[0]))

	// check that the interval is returned by AllIntervals
	intervals = top.AllIntervals()
	Tassert(t, len(intervals) == 3, "Expected 3 intervals, got %d", len(intervals))
	Tassert(t, intervals[1].Equal(iv), fmt.Sprintf("Expected %v, got %v", iv, intervals[1]))

	tree.Verify(t, top, false, false)

}
