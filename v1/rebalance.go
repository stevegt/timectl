package timectl

// rebalance checks and corrects the balance of the tree using DSW algorithm.
func (old *Tree) rebalance() (t *Tree) {
	t = old
	if t == nil {
		return nil
	}

	// Step 1: Convert the tree to a vine.
	t = t.vine()

	// Step 2: Convert the vine to a balanced tree.
	t = t.balance()

	return
}

// vine converts the tree to a vine.
func (t *Tree) vine() (newRoot *Tree) {
	// XXX checkpoint before converting to store values in internal nodes
}

// rotateLeft performs a left rotation on this node.
func (t *Tree) rotateLeft() (newRoot *Tree) {
	y := t.right
	t2 := y.left

	y.left = t
	t.right = t2

	newRoot = y
	return
}

// rotateRight performs a right rotation on this node.
func (t *Tree) rotateRight() (newRoot *Tree) {
	y := t.left
	t2 := y.right

	y.right = t
	t.left = t2

	newRoot = y
	return
}

// getBalance calculates and returns the balance factor of this node.
func (t *Tree) getBalance() int {
	if t == nil {
		return 0
	}
	return t.left.height() - t.right.height()
}
