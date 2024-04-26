package timectl

// Delete removes an interval from the tree, adjusting the structure as
// necessary. Deletion fails if the interval is not found in the tree.
func (t *Tree) Delete(interval Interval) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.delete(interval)
}

// delete is a non-threadsafe version of Delete for internal use.
func (t *Tree) delete(interval Interval) bool {
	if t == nil {
		return false
	}

	if t.leafInterval != nil && t.leafInterval.Equal(interval) {
		if t.left == nil && t.right == nil {
			// If node is a leaf
			t.leafInterval = nil
			return true
		}

		// If the node has one child, replace this node with the child
		if t.left == nil {
			*t = *t.right
			return true
		}
		if t.right == nil {
			*t = *t.left
			return true
		}

		// Node has two children, find the in-order successor
		successor := t.right.min()
		t.leafInterval = successor.leafInterval
		return t.right.delete(t.leafInterval) // Delete the successor
	}

	// Search in children
	if t.left != nil && t.left.overlaps(interval) {
		if t.left.delete(interval) {
			if t.left.leafInterval == nil && t.left.left == nil && t.left.right == nil {
				t.left = nil // Prune empty child
			}
			return true
		}
	}
	if t.right != nil && t.right.overlaps(interval) {
		if t.right.delete(interval) {
			if t.right.leafInterval == nil && t.right.left == nil && t.right.right == nil {
				t.right = nil // Prune empty child
			}
			return true
		}
	}

	return false
}

// overlaps checks if the tree's interval overlaps with the given interval.
func (t *Tree) overlaps(interval Interval) bool {
	if t.leafInterval != nil {
		return t.leafInterval.Conflicts(interval, true)
	}
	return false
}

// isEmpty checks if the tree node is empty, which is true if it has no interval and no children.
func (t *Tree) isEmpty() bool {
	return t.leafInterval == nil && t.left == nil && t.right == nil
}

// min finds the minimum interval in the subtree rooted at the current node.
func (t *Tree) min() *Tree {
	curr := t
	for curr.left != nil {
		curr = curr.left
	}
	return curr
}
