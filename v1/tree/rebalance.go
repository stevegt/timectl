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
	// t = t.balance()

	return
}

// vine converts the tree to a vine.
func (t *Tree) vine() (newRoot *Tree) {
	// XXX checkpoint before converting to store values in internal nodes
	return t
}

// rotateLeft performs a left rotation on this node.
func (t *Tree) rotateLeft() (newRoot *Tree) {
	newRoot = t.right
	newLeftRight := t.right.left
	newRoot.left = t
	newRoot.left.right = newLeftRight
	newRoot.left.setMinMax()
	newRoot.right.setMinMax()
	newRoot.setMinMax()
	return
}

// rotateRight performs a right rotation on this node.
func (t *Tree) rotateRight() (newRoot *Tree) {
	newRoot = t.left
	newRightLeft := t.left.right
	newRoot.right = t
	newRoot.right.left = newRightLeft
	newRoot.left.setMinMax()
	newRoot.right.setMinMax()
	newRoot.setMinMax()
	return
}

// setMinMax updates the minimum and maximum values of this node.
func (t *Tree) setMinMax() {
	if t == nil {
		return
	}
	if t.left == nil {
		t.minStart = t.interval.Start()
	} else {
		t.minStart = t.left.minStart
	}
	if t.right == nil {
		t.maxEnd = t.interval.End()
	} else {
		t.maxEnd = t.right.maxEnd
	}
	t.setMaxPriority()
}

// getBalance calculates and returns the balance factor of this node.
func (t *Tree) getBalance() int {
	if t == nil {
		return 0
	}
	return t.left.height() - t.right.height()
}
