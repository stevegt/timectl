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
	if t == nil || (t.leafInterval == nil && t.left == nil && t.right == nil) {
		return false
	}

	/*
		// Direct match of the interval in the current node
		if t.leafInterval != nil && t.leafInterval.Equal(interval) {
			// Node with two children case
			if t.left != nil && t.right != nil {
				successor := t.right.min()
				t.leafInterval = successor.leafInterval
				return t.right.delete(successor.leafInterval) // Recursively delete the successor
			} else if t.left != nil { // Node with only left child
				*t = *t.left
				return true
			} else if t.right != nil { // Node with only right child
				*t = *t.right
				return true
			} else { // Leaf node
				t.leafInterval = nil // Clear the interval
				return true
			}
		}
	*/

	// Recursively check left and right subtree for the interval
	deletedLeft := false
	if t.left != nil && t.left.overlaps(interval) {
		deletedLeft = t.left.delete(interval)
	}

	deletedRight := false
	if t.right != nil && t.right.overlaps(interval) {
		deletedRight = t.right.delete(interval)
	}

	// If an interval was successfully deleted from either subtree, cleanup subtree if needed
	if deletedLeft && t.left != nil && t.left.isEmpty() {
		t.left = nil // Left cleanup
	}
	if deletedRight && t.right != nil && t.right.isEmpty() {
		t.right = nil // Right cleanup
	}

	return deletedLeft || deletedRight
}

// FindExact searches for an exact match of the given interval in the tree.
// If found, it returns the matching interval and its parent node within the tree.
// The return value parent is nil if the found interval is at the root.
func (t *Tree) FindExact(interval Interval) (*Tree, *Tree) {
	if t == nil {
		return nil, nil
	}

	var parent *Tree = nil
	current := t

	for {
		if current.leafInterval != nil && current.leafInterval.Equal(interval) {
			return current, parent
		}

		if current.left != nil && current.left.overlaps(interval) {
			parent = current
			current = current.left
		} else if current.right != nil && current.right.overlaps(interval) {
			parent = current
			current = current.right
		} else {
			return nil, parent
		}
	}
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
