package tree

import (
	"math"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// Delete removes an interval from the tree
func (t *Node) Delete(interval interval.Interval) (out *Node, err error) {
	defer Return(&err)
	t.mu.Lock()
	defer t.mu.Unlock()

	_, found := t.findExact(interval, nil)
	Assert(found != nil, "Interval not found: %v", interval)

	// Free the node, discarding the old interval.
	_ = t.free(found)

	// merge free siblings
	out = t.mergeFree()

	return
}

// free sets the interval of the node to a free interval and updates
// the min/max values.  The node's old interval is still intact, but
// no longer part of the tree.  We return the old interval so that the
// caller can decide what to do with it.
func (t *Node) free(node *Node) (old interval.Interval) {
	old = node.Interval()
	freeInterval := interval.NewInterval(node.Start(), node.End(), 0)
	node.SetInterval(freeInterval)
	return
}

// RemoveRange removes all intervals that start or end within the
// given time range.  It returns the removed intervals.  It does not
// return intervals that are marked as free (priority 0) -- it
// instead adjusts free intervals to fill gaps in the tree.
func (t *Node) RemoveRange(start, end time.Time) (out *Node, removed []interval.Interval) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Find all nodes that start or end within the given range.
	duration := end.Sub(start)
	nodes := t.FindLowerPriority(true, start, end, duration, math.MaxFloat64)

	// free the nodes' intervals
	// XXX refactor FindLowerPriority to return a tree instead of a slice
	freed := make([]interval.Interval, len(nodes))
	for i, n := range nodes {
		freed[i] = t.free(n)
	}

	// only return non-free intervals
	for _, iv := range freed {
		if iv.Priority() == 0 {
			continue
		}
		removed = append(removed, iv)
	}

	// merge free siblings
	out = t.mergeFree()

	return
}
