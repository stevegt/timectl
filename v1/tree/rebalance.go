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

// rotateLeft performs a Left rotation on this node.
func (t *Tree) rotateLeft() (newRoot *Tree) {
	newRoot = t.Right
	newLeftRight := t.Right.Left
	newRoot.Left = t
	newRoot.Left.Right = newLeftRight
	newRoot.Left.setMinMax()
	newRoot.Right.setMinMax()
	newRoot.setMinMax()
	return
}

// rotateRight performs a Right rotation on this node.
func (t *Tree) rotateRight() (newRoot *Tree) {
	newRoot = t.Left
	newRightLeft := t.Left.Right
	newRoot.Right = t
	newRoot.Right.Left = newRightLeft
	newRoot.Left.setMinMax()
	newRoot.Right.setMinMax()
	newRoot.setMinMax()
	return
}

// setMinMax updates the minimum and maximum values of this node.
func (t *Tree) setMinMax() {
	if t == nil {
		return
	}
	if t.Left == nil {
		t.MinStart = t.Interval.Start()
	} else {
		t.MinStart = t.Left.MinStart
	}
	if t.Right == nil {
		t.MaxEnd = t.Interval.End()
	} else {
		t.MaxEnd = t.Right.MaxEnd
	}
	t.setMaxPriority()
}

// getBalance calculates and returns the balance factor of this node.
func (t *Tree) getBalance() int {
	if t == nil {
		return 0
	}
	return t.Left.height() - t.Right.height()
}
