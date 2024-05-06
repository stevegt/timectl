package tree

import (
	"fmt" // Import the fmt package to use for formatting errors.
	"time"

	"github.com/stevegt/timectl/util"
	// . "github.com/stevegt/goadapt"
)

// Verify checks the integrity of the tree structure. It makes sure
// that all nodes and intervals are correctly placed within the tree
// according to the interval tree properties.
func (t *Node) Verify(ckBalance bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// the interval should not be nil
	if t.Interval() == nil {
		return fmt.Errorf("root interval is nil")
	}

	// - the first interval in the tree should be a free interval that
	//   starts at TreeStart
	firstInterval := t.FirstNode().Interval()
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
	lastInterval := t.LastNode().Interval()
	if lastInterval == nil {
		return fmt.Errorf("last interval is nil")
	}
	if !lastInterval.End().Equal(TreeEnd) {
		return fmt.Errorf("last interval end time does not match tree end time")
	}
	if lastInterval.Busy() {
		return fmt.Errorf("last interval is not free")
	}

	var prevNode *Node
	iter := NewIterator(t, true)
	for {
		node := iter.Next()
		if node == nil {
			break
		}
		// Pf(" got: %-10s %v\n", path, node.leafInterval)

		// the node interval should not be nil
		if node.Interval() == nil {
			return fmt.Errorf("node interval is nil")
		}

		// the node's parent should not be nil unless it is the root
		if node.Parent() == nil && node != t {
			return fmt.Errorf("node parent is nil")
		}

		// the node should be a child of its parent
		if node.Parent() != nil && node.Parent().Left() != node && node.Parent().Right() != node {
			return fmt.Errorf("node is not a child of its parent")
		}

		start := node.Interval().Start()
		end := node.Interval().End()

		// - each interval's end time should be greater than its start time
		if !end.After(start) {
			return fmt.Errorf("interval end time is not after start time")
		}

		// - each interval's minStart time should be equal to the minimum
		//   of its start time and the start time of its left child
		expectMinStart := node.MinStart()
		if node.Left() != nil {
			gotMinStart := util.MinTime(start, node.Left().MinStart())
			if !expectMinStart.Equal(gotMinStart) {
				return fmt.Errorf("%s minStart time does not match minimum of start time and left child minStart time", node)
			}
		} else {
			if !expectMinStart.Equal(start) {
				return fmt.Errorf("%s minStart time does not match interval start time", node)
			}
		}

		// - each interval's maxEnd time should be equal to the maximum
		//   of its end time and the end time of its right child
		expectMaxEnd := node.MaxEnd()
		if node.Right() != nil {
			gotMaxEnd := util.MaxTime(end, node.Right().MaxEnd())
			if !expectMaxEnd.Equal(gotMaxEnd) {
				return fmt.Errorf("%s maxEnd time does not match maximum of end time and right child maxEnd time", node)
			}
		} else {
			if !expectMaxEnd.Equal(end) {
				return fmt.Errorf("%s maxEnd time does not match interval end time", node)
			}
		}

		if prevNode != nil {
			// - each interval's start time should be equal to the end time
			//   of the previous interval
			ad := util.AbsDuration(prevNode.Interval().End().Sub(start))
			if ad > time.Second {
				return fmt.Errorf("start time does not match previous interval end time: %v\nprev: %s\nnode: %s", ad, prevNode, node)
			}

			// - there should be no adjacent free intervals
			if !prevNode.Busy() && !node.Busy() {
				return fmt.Errorf("adjacent free intervals")
			}
		}
		prevNode = node

	}

	if ckBalance {
		err := t.ckBalance(nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// ckBalance checks the balance of the tree.
func (t *Node) ckBalance(ancestors Path) error {
	if t == nil {
		return nil
	}
	myPath := ancestors.Append(t)

	// check this node's balance
	leftHeight := t.Left().CalcHeight()
	rightHeight := t.Right().CalcHeight()
	if leftHeight < rightHeight-1 || rightHeight < leftHeight-1 {
		return fmt.Errorf("path: %v left height: %d, right height: %d", myPath, leftHeight, rightHeight)
	}

	if false {
		//   check the balance of the left and right subtrees
		leftBalance := t.Left().ckBalance(myPath)
		if leftBalance != nil {
			return leftBalance
		}
		rightBalance := t.Right().ckBalance(myPath)
		if rightBalance != nil {
			return rightBalance
		}
	}

	return nil
}

// height returns the height of the tree.
func (t *Node) CalcHeight() int {
	if t == nil {
		return 0
	}
	return 1 + max(t.Left().CalcHeight(), t.Right().CalcHeight())
}
