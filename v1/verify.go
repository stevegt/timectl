package timectl

import (
	"fmt" // Import the fmt package to use for formatting errors.
	// . "github.com/stevegt/goadapt"
)

// Verify checks the integrity of the tree structure. It makes sure
// that all nodes and intervals are correctly placed within the tree
// according to the interval tree properties.
func (t *Tree) Verify() error {

	// - the root node should span the entire range from TreeStart to
	// TreeEnd
	rootStart := t.Interval().Start()
	rootEnd := t.Interval().End()
	if !rootStart.Equal(TreeStart) {
		return fmt.Errorf("root interval start time does not match tree start time")
	}
	if !rootEnd.Equal(TreeEnd) {
		return fmt.Errorf("root interval end time does not match tree end time")
	}

	// - the first interval in the tree should be a free interval that
	//   starts at TreeStart
	firstInterval := t.firstNode().leafInterval
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
	lastInterval := t.lastNode().leafInterval
	if lastInterval == nil {
		return fmt.Errorf("last interval is nil")
	}
	if !lastInterval.End().Equal(TreeEnd) {
		return fmt.Errorf("last interval end time does not match tree end time")
	}
	if lastInterval.Busy() {
		return fmt.Errorf("last interval is not free")
	}

	var prevLeaf *Tree
	for path := range t.allPaths(nil) {
		node := path.Last()
		// Pf(" got: %-10s %v\n", path, node.leafInterval)

		// - each node should have either two children or none
		if t.left == nil && t.right != nil {
			return fmt.Errorf("node has right child but no left child")
		}
		if t.left != nil && t.right == nil {
			return fmt.Errorf("node has left child but no right child")
		}

		if node.leafInterval == nil {
			continue
		}
		// only leaf interval checks below here

		// - each interval's end time should be greater than its start time
		if !node.End().After(node.Start()) {
			return fmt.Errorf("interval end time is not after start time")
		}

		if prevLeaf != nil {
			// - each interval's start time should be equal to the end time
			//   of the previous interval
			if !node.Start().Equal(prevLeaf.End()) {
				return fmt.Errorf("%s start time does not match previous interval end time", path)
			}

			// - there should be no adjacent free intervals
			if !prevLeaf.Busy() && !node.Busy() {
				return fmt.Errorf("adjacent free intervals")
			}
		}
		prevLeaf = node

	}

	return nil
}
