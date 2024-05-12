package tree

import (
	"math"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/v2/interval"
)

// Delete removes an interval from the tree
func (t *Node) Delete(interval interval.Interval) (out *Node, err error) {
	defer Return(&err)
	t.mu.Lock()
	defer t.mu.Unlock()

	path := t.findExact(interval, nil)
	found := path.Last()
	Assert(found != nil, "Interval not found: %v", interval)

	return t.deleteNode(found)
}

// deleteNode removes a node from the tree
func (t *Node) deleteNode(node *Node) (out *Node, err error) {
	// XXX should be:
	// out = t.clone()

	// Free the node, discarding the old interval.
	_ = node.free()

	// merge free siblings
	out = t.mergeFree()
	// XXX should be:
	// out = out.mergeFree()

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
	nodes, _ := t.FindLowerPriority(true, start, end, duration, math.MaxFloat64)

	// free the nodes' intervals
	// XXX refactor FindLowerPriority to return a tree instead of a slice
	freed := make([]interval.Interval, len(nodes))
	for i, n := range nodes {
		freed[i] = n.Interval()
		_ = n.free()
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
