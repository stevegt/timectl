package tree

// rebalance checks and corrects the balance of the tree using DSW algorithm.
func (old *Tree) rebalance() (t *Tree) {
	t = old
	if t == nil {
		return nil
	}

	// Step 1: Convert the tree to a vine.
	t = t.vine()

	// Step 2: Convert the vine to a balanced tree.
	// t = t.balance()

	return
}

// vine converts the tree to a vine.
func (t *Tree) vine() (newRoot *Tree) {
	// XXX checkpoint before converting to store values in internal nodes
	return t
}

// getBalance calculates and returns the balance factor of this node.
func (t *Tree) getBalance() int {
	if t == nil {
		return 0
	}
	return t.Left.height() - t.Right.height()
}
