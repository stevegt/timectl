package timectl

// mergeFree merges adjacent free intervals in the tree.
func (t *Tree) mergeFree() {
	if t.left != nil && t.right != nil && !t.left.leafInterval.Busy() && !t.right.leafInterval.Busy() {
		// Merge left and right free intervals into a single interval.
		t.leafInterval = NewInterval(t.left.leafInterval.Start(), t.right.leafInterval.End(), 0)
		t.left = nil  // Clear left child.
		t.right = nil // Clear right child.
	}
}
