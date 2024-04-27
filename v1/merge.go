package timectl

// mergeFree merges adjacent free intervals in the tree.
func (t *Tree) mergeFree() {
	// Check for nil because this function could be called on a nil receiver due to defer in Tree operations.
	if t == nil {
		return
	}

	// Lock the tree for safe concurrent access.
	t.mu.Lock()
	defer t.mu.Unlock()

	// Recursive helper function to merge free intervals starting at a given node.
	var mergeRecursive func(node *Tree)
	mergeRecursive = func(node *Tree) {
		if node == nil || node.leafInterval == nil {
			// If node is nil or not a leaf, continue to its children.
			if node.left != nil {
				mergeRecursive(node.left)
			}
			if node.right != nil {
				mergeRecursive(node.right)
			}
		} else {
			// When a leaf node with a free interval is found, attempt to merge it with adjacent nodes.
			if node.left != nil && !node.left.interval().Busy() {
				// Merge with left child if it's a free interval.
				node.leafInterval = NewInterval(node.left.interval().Start(), node.leafInterval.End(), 0)
				node.left = nil // Remove the left node after merging.
			}
			if node.right != nil && !node.right.interval().Busy() {
				// Merge with right child if it's a free interval.
				node.leafInterval = NewInterval(node.leafInterval.Start(), node.right.interval().End(), 0)
				node.right = nil // Remove the right node after merging.
			}
		}
	}

	// Start the merging process from the root of the tree.
	mergeRecursive(t)
}
