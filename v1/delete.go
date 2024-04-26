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
	if t.leafInterval != nil {
		if t.leafInterval.Equal(interval) {
			// If the current node is a leaf with the matching interval, remove the interval and clean up the node.
			t.leafInterval = nil
			t.left = nil
			t.right = nil
			return true
		}
		return false // The leaf node's interval does not match.
	}

	// If the interval matches exactly the node's interval (which could be a synthetic interval encompassing child nodes),
	// and both child nodes are leaf or nil, we can collapse this node.
	if t.interval().Equal(interval) && (t.left == nil || t.left.isLeaf()) && (t.right == nil || t.right.isLeaf()) {
		t.leafInterval = nil // Clear the interval assuming it's a synthetic interval covering the whole node.
		if t.left != nil && t.left.leafInterval != nil {
			t.promoteChild(t.left) // Promote the left child
		} else if t.right != nil && t.right.leafInterval != nil {
			t.promoteChild(t.right) // Promote the right child
		}
		return true
	}

	foundInLeft, foundInRight := false, false
	if t.left != nil {
		foundInLeft = t.left.delete(interval)
	}
	if t.right != nil {
		foundInRight = t.right.delete(interval)
	}

	// After a deletion, check if either child became empty and remove it.
	if foundInLeft || foundInRight {
		if t.left != nil && t.left.isEmpty() {
			t.left = nil
		}
		if t.right != nil && t.right.isEmpty() {
			t.right = nil
		}
	}

	if !foundInLeft && !foundInRight {
		// The interval was not found in either subtree.
		return false
	}

	// Attempt to balance or simplify the tree after deletion if necessary.
	t.balanceOrSimplify()

	return true
}

// isEmpty checks if the tree node is empty, which is true if it has no interval and no children.
func (t *Tree) isEmpty() bool {
	return t.leafInterval == nil && t.left == nil && t.right == nil
}

// balanceOrSimplify tries to simplify the tree structure after a deletion
// by either removing unnecessary nodes or balancing the tree.
func (t *Tree) balanceOrSimplify() {
	if t.isLeaf() {
		return // Nothing to simplify or balance if it's a leaf node.
	}

	// Check if either child is nil and promote the other.
	if t.left == nil && t.right != nil {
		t.promoteChild(t.right)
	} else if t.right == nil && t.left != nil {
		t.promoteChild(t.left)
	}

	// Post-promotion, if the current node becomes a leaf node, attempt further simplification.
	if t.isLeaf() {
		t.balanceOrSimplify() // Further checks if simplification is possible.
	}
}

// promoteChild replaces the current tree node with the child node.
func (t *Tree) promoteChild(child *Tree) {
	t.leafInterval = child.leafInterval
	t.left = child.left
	t.right = child.right
}

// isLeaf checks if the current tree node is a leaf (i.e., has an interval and no children).
func (t *Tree) isLeaf() bool {
	return t.leafInterval != nil && t.left == nil && t.right == nil
}

