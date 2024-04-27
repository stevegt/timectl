package timectl

// mergeFree merges adjacent free intervals in the tree.
func (t *Tree) mergeFree() {
	// Check for nil because this function could be called on a nil receiver due to defer in Tree operations.
	if t == nil {
		return
	}

	// Recursive helper function to merge free intervals starting at a given node.
	var mergeRecursive func(node *Tree)
	mergeRecursive = func(node *Tree) {
		if node.left != nil && node.right != nil && !node.left.busy() && !node.right.busy() {
			// Merge the two free intervals.
			leftStart := node.left.treeStart()
			rightEnd := node.right.treeEnd()
			node.interval = NewInterval(leftStart, rightEnd, 0)
			node.left = nil
			node.right = nil
		} else {
			if node.left != nil {
				// Continue merging free intervals on the left.
				mergeRecursive(node.left)
			}
			if node.right != nil {
				// Continue merging free intervals on the right.
				mergeRecursive(node.right)
			}
		}
	}

	// Start the merging process from the root of the tree.
	mergeRecursive(t)
}
