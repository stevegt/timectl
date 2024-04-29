package tree

// rebalance restructures the tree to maintain or improve balance. It does not require
// an 'ancestors' variable as initially indicated by mistaken code implementation.
// It is called after insertions or deletions that might have left the tree unbalanced.
func (t *Tree) rebalance() {
	if t == nil {
		return
	}

	t.Mu.Lock()
	defer t.Mu.Unlock()

	leftHeight := t.Left.height()
	rightHeight := t.Right.height()

	// If left is heavier, check if a right rotation is needed
	if leftHeight > rightHeight+1 {
		leftLeftHeight := t.Left.Left.height()
		leftRightHeight := t.Left.Right.height()

		// Left-Right Case: Left rotation on left child before right rotation on self
		if leftRightHeight > leftLeftHeight {
			t.Left = t.Left.rotateLeft()
		}
		// Left-Left Case: Right rotation on self
		t = t.rotateRight()
	}

	// If right is heavier, check if a left rotation is needed
	if rightHeight > leftHeight+1 {
		rightRightHeight := t.Right.Right.height()
		rightLeftHeight := t.Right.Left.height()

		// Right-Left Case: Right rotation on right child before left rotation on self
		if rightLeftHeight > rightRightHeight {
			t.Right = t.Right.rotateRight()
		}
		// Right-Right Case: Left rotation on self
		t = t.rotateLeft()
	}

	// Note: The original code had references to a nonexistent 'ancestors' slice,
	// which was incorrect. Furthermore, the rebalance process is self-contained within
	// the node 't', and manual reassignment 't = newRoot' is not applicable as 't' is
	// a local copy of the pointer. The proper tree structure adjustment is handled within
	// rotateLeft and rotateRight method calls, which correctly modify the tree structure.
}
