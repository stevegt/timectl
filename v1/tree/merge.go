package tree

import "github.com/stevegt/timectl/interval"

// MergeFree merges adjacent free intervals in the tree.
func (t *Tree) MergeFree() {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	// Check for nil because this function could be called on a nil receiver due to defer in Tree operations.
	if t == nil {
		return
	}

	// Recursive helper function to merge free intervals starting at a given node.
	var mergeRecursive func(node *Tree)
	mergeRecursive = func(node *Tree) {
		if node.Left != nil && node.Right != nil && !node.Left.Busy() && !node.Right.Busy() {
			// Merge the two free intervals.
			leftStart := node.Left.MinStart
			rightEnd := node.Right.MaxEnd
			node.Interval = interval.NewInterval(leftStart, rightEnd, 0)
			node.Left = nil
			node.Right = nil
		} else {
			if node.Left != nil {
				// Continue merging free intervals on the left.
				mergeRecursive(node.Left)
			}
			if node.Right != nil {
				// Continue merging free intervals on the right.
				mergeRecursive(node.Right)
			}
		}
	}

	// Start the merging process from the root of the tree.
	mergeRecursive(t)
}
