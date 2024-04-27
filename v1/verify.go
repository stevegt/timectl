package timectl

import (
	"fmt" // Import the fmt package to use for formatting errors.
	// . "github.com/stevegt/goadapt"
)

// Verify checks the integrity of the tree structure. It makes sure
// that all nodes and intervals are correctly placed within the tree
// according to the interval tree properties.
func (t *Tree) Verify() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// the interval should not be nil
	if t.interval == nil {
		return fmt.Errorf("root interval is nil")
	}

	// - the first interval in the tree should be a free interval that
	//   starts at TreeStart
	firstInterval := t.firstNode().interval
	if firstInterval == nil {
		return fmt.Errorf("first interval is nil")
	}
	if !firstInterval.Start().Equal(TreeStart) {
		return fmt.Errorf("first interval start time does not match tree start time")
	}
	if firstInterval.Busy() {
		return fmt.Errorf("first interval is not free")
	}

	// - the last interval in the tree should be a free interval that
	//   ends at TreeEnd
	lastInterval := t.lastNode().interval
	if lastInterval == nil {
		return fmt.Errorf("last interval is nil")
	}
	if !lastInterval.End().Equal(TreeEnd) {
		return fmt.Errorf("last interval end time does not match tree end time")
	}
	if lastInterval.Busy() {
		return fmt.Errorf("last interval is not free")
	}

	var prevNode *Tree
	for path := range t.allPaths(nil) {
		node := path.Last()
		// Pf(" got: %-10s %v\n", path, node.leafInterval)

		// the node interval should not be nil
		if node.interval == nil {
			return fmt.Errorf("node interval is nil")
		}

		start := node.interval.Start()
		end := node.interval.End()

		// - each interval's end time should be greater than its start time
		if !end.After(start) {
			return fmt.Errorf("interval end time is not after start time")
		}

		// - each interval's minStart time should be equal to the minimum
		//   of its start time and the start time of its left child
		expectMinStart := node.minStart
		if node.left != nil {
			gotMinStart := MinTime(start, node.left.minStart)
			if !expectMinStart.Equal(gotMinStart) {
				return fmt.Errorf("%s minStart time does not match minimum of start time and left child minStart time", path)
			}
		} else {
			if !expectMinStart.Equal(start) {
				return fmt.Errorf("%s minStart time does not match interval start time", path)
			}
		}

		// - each interval's maxEnd time should be equal to the maximum
		//   of its end time and the end time of its right child
		expectMaxEnd := node.maxEnd
		if node.right != nil {
			gotMaxEnd := MaxTime(end, node.right.maxEnd)
			if !expectMaxEnd.Equal(gotMaxEnd) {
				return fmt.Errorf("%s maxEnd time does not match maximum of end time and right child maxEnd time", path)
			}
		} else {
			if !expectMaxEnd.Equal(end) {
				return fmt.Errorf("%s maxEnd time does not match interval end time", path)
			}
		}

		if prevNode != nil {
			// - each interval's start time should be equal to the end time
			//   of the previous interval
			if !start.Equal(prevNode.interval.End()) {
				return fmt.Errorf("%s start time does not match previous interval end time", path)
			}

			// - there should be no adjacent free intervals
			if !prevNode.Busy() && !node.Busy() {
				return fmt.Errorf("adjacent free intervals")
			}
		}
		prevNode = node

	}

	/*
		err := t.ckBalance(nil)
		if err != nil {
			return err
		}
	*/

	return nil
}

// ckBalance checks the balance of the tree. It makes sure that the
// tree is balanced according to the AVL tree properties.
func (t *Tree) ckBalance(ancestors Path) error {
	if t == nil {
		return nil
	}
	myPath := ancestors.Append(t)

	// - the height of the left and right subtrees of every node differ
	//   by at most 1
	leftHeight := t.left.height()
	rightHeight := t.right.height()
	if leftHeight < rightHeight-1 || rightHeight < leftHeight-1 {
		return fmt.Errorf("height of left and right subtrees of %s differ by more than 1", myPath)
	}

	// - the height of the left and right subtrees of every node differ
	//   by at most 1
	leftBalance := t.left.ckBalance(myPath)
	if leftBalance != nil {
		return leftBalance
	}
	rightBalance := t.right.ckBalance(myPath)
	if rightBalance != nil {
		return rightBalance
	}

	return nil
}

// height returns the height of the tree.
func (t *Tree) height() int {
	if t == nil {
		return 0
	}
	return 1 + max(t.left.height(), t.right.height())
}
