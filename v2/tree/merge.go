package tree

import "github.com/stevegt/timectl/v2/interval"

// mergeFree merges adjacent free intervals in the tree and returns a
// vine with the merged intervals.
func (t *Node) mergeFree() (vine *Node) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t == nil {
		return
	}

	// turn the tree into a vine and merge free intervals
	vine, _ = t.treeToVine()

	// merge free intervals
	node := vine
	for {
		if node.Right() == nil {
			break
		}
		if !node.Busy() && !node.Right().Busy() {
			node.SetInterval(interval.NewInterval(node.Interval().Start(), node.Right().Interval().End(), 0))
			node.SetRight(node.Right().Right())
			// see if we can merge more intervals with this node
			continue
		}
		// advance node
		node = node.Right()
	}

	return
}
