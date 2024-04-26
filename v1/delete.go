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
		t.leafInterval = nil
		t.left = nil
		t.right = nil
		return true
	}

	if t.left != nil && t.left.interval().Conflicts(interval, true) {
		if t.left.delete(interval) {
			if t.left.isEmpty() {
				t.left = nil
			}
			return true
		}
	}
	if t.right != nil && t.right.interval().Conflicts(interval, true) {
		if t.right.delete(interval) {
			if t.right.isEmpty() {
				t.right = nil
			}
			return true
		}
	}

	return false
}

// isEmpty checks if the tree node is empty, which is true if it has no interval and no children.
func (t *Tree) isEmpty() bool {
	return t.leafInterval == nil && t.left == nil && t.right == nil
}
