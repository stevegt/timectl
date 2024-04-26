package timectl

// . "github.com/stevegt/goadapt"

// Delete removes an interval from the tree, adjusting the structure as
// necessary.  Deletion fails if the interval is not found in the tree.
func (t *Tree) Delete(interval Interval) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.delete(interval)
}

// delete is a non-threadsafe version of Delete for internal use.
func (t *Tree) delete(interval Interval) bool {
	// check leaf node for interval
	if t.leafInterval != nil && t.leafInterval.Equal(interval) {
		t.leafInterval = nil
		t.left = nil
		t.right = nil
		return true
	}

	// check children for interval
	found := false
	for _, child := range []*Tree{t.left, t.right} {
		if child != nil && child.Delete(interval) {
			// interval was found and deleted in child or descendant
			found = true
			if child.leafInterval == nil && child.left == nil && child.right == nil {
				// child is a leaf node with no children, so remove it
				if child == t.left {
					t.left = nil
				}
				if child == t.right {
					t.right = nil
				}
			}
		}
	}
	if !found {
		return false
	}

	// see if we can promote a child
	if t.leafInterval == nil && t.left == nil && t.right != nil {
		// promote right child
		t.leafInterval = t.right.leafInterval
		t.left = t.right.left
		t.right = t.right.right
	}
	if t.leafInterval == nil && t.right == nil && t.left != nil {
		// promote left child
		t.leafInterval = t.left.leafInterval
		t.left = t.left.left
		t.right = t.left.right
	}

	// see if we can merge children
	if t.left != nil && t.right != nil {
		if !t.left.Busy() && !t.right.Busy() {
			// both children are free, so replace them with a single free node
			t.leafInterval = NewInterval(t.left.Start(), t.right.End(), 0)
			t.left = nil
			t.right = nil
		}
	}

	return true
}
