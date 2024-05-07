package tree

import (
	"math"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// Delete removes an interval from the tree and returns the modified tree.
func (t *Node) Delete(interval interval.Interval) (out *Node, err error) {
	defer Return(&err)
	t.mu.Lock()
	defer t.mu.Unlock()

	_, found := t.findExact(interval, nil)
	Assert(found != nil, "Interval not found: %v", interval)

	// Free the node, discarding the old interval.
	out = t.free2(found)

	// merge free siblings
	out = t.mergeFree()

	return
}

// free sets the interval of the node to a free interval, updates
// the min/max values.  The node's old interval is still intact, but
// no longer part of the tree.  We return a new tree.
func (t *Tree) free2(node *Node) (out *Tree) {
	out = t.clone()

	freeInterval := interval.NewInterval(node.Start(), node.End(), 0)
	node.SetInterval(freeInterval)
	return
}

// RemoveRange removes all nodes whose intervals start or end within
// the given time range.  It returns a modified tree and the removed
// intervals.  It does not return intervals that are marked as free
// (priority 0) -- it instead adjusts free intervals to fill gaps in
// the tree.
// XXX move to Tree
func (t *Node) RemoveRange(start, end time.Time) (out *Node, removed []interval.Interval) {
	t.mu.Lock()
	defer t.mu.Unlock()

	out = t.clone()

	// Find all nodes that start or end within the given range.
	duration := end.Sub(start)
	found, _ := out.FindLowerPriority(true, start, end, duration, math.MaxFloat64, nil)

	// free the nodes' intervals
	iter := NewIterator(found, true)
	for {
		path := iter.Next()
		node := path.Last()
		if node == nil {
			break
		}
		if node.Priority() == 0 {
			continue
		}
		removed = append(removed, node.Interval())
		out.free(node)
	}

	// merge free siblings
	out = out.mergeFree()

	return
}
