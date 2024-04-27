package timectl

import (
	"sync"
)

// Tree represents a node in an interval tree.
type Tree struct {
	// If this is not a leaf node, leafInterval is nil.
	leafInterval Interval
	left         *Tree // Pointer to the left child
	right        *Tree // Pointer to the right child

	mu sync.RWMutex
}

// Delete removes an interval from the tree and returns true if the interval was successfully removed.
func (t *Tree) Delete(interval Interval) (ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.delete(interval)
}

// delete is the internal recursive method to remove an interval.
func (t *Tree) delete(interval Interval) (ok bool) {
	if t.leafInterval != nil && t.leafInterval.Equal(interval) {
		*t = Tree{} // Clear the leaf node creating an empty tree node.
		return true
	}

	if t.left != nil {
		if t.left.delete(interval) {
			if t.right == nil {
				*t = *t.left // Promote the left child.
			} else {
				t.mergeFree()
			}
			return true
		}
	}
	if t.right != nil {
		if t.right.delete(interval) {
			if t.left == nil {
				*t = *t.right // Promote the right child.
			} else {
				t.mergeFree()
			}
			return true
		}
	}
	return false
}

// mergeFree merges adjacent free intervals in the tree.
func (t *Tree) mergeFree() {
	if t.left != nil && t.right != nil && !t.left.leafInterval.Busy() && !t.right.leafInterval.Busy() {
		// Merge left and right free intervals into a single interval.
		t.leafInterval = NewInterval(t.left.leafInterval.Start(), t.right.leafInterval.End(), 0)
		t.left = nil  // Clear left child.
		t.right = nil // Clear right child.
	}
}
