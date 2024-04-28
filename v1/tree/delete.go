package tree

import (
	"fmt"

	"github.com/stevegt/timectl/interval"
)

// Delete removes an interval from the tree and returns true if the interval was successfully removed.
func (t *Tree) Delete(interval interval.Interval) bool {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	path, found := t.findExact(interval, nil)
	if found == nil {
		// interval.Interval not found.
		return false
	}

	// Proceed to delete the node.
	return t.delete(path, found)
}

func (t *Tree) free(node *Tree) error {
	if node.Left != nil || node.Right != nil {
		return fmt.Errorf("cannot free node with children")
	}
	freeInterval := interval.NewInterval(node.Start(), node.End(), 0)
	node.Interval = freeInterval
	return nil
}

func (t *Tree) delete(path []*Tree, node *Tree) bool {
	// Placeholder implementation.
	return false
}

/*
// delete is simply a process of finding and freeing the target
// interval and then merging with free siblings.
func (t *Tree) delete(interval interval.Interval) (deletedInterval interval.Interval) {

	// Find the node containing the interval.
	ancestors, node := t.findExact(interval, nil)

	// If the node is not found, return nil.
	if node == nil {
		return nil
	}

	// Free the node.
	if err := t.free(node); err != nil {
		return nil
	}

	// Merge with free siblings.
	tree.mergeFreeSiblings(ancestors)

	// If node is a root node with no children, just remove the interval.
	if len(path) == 0 {
		if node.left == nil && node.right == nil {
			t.leafInterval = nil
			return true
		}
	}

	// Handle deletion logic when the node is not the root.
	parent := path[len(path)-1] // Get the parent node.

	if parent.left == node {
		parent.left = nil // If the node is a left child.
	} else if parent.right == node {
		parent.right = nil // If the node is a right child.
	}

	// After deletion, check if the parent node can absorb the interval of the deleted node.
	t.tryAbsorb(parent)

	// Check if there's a need to rebalance or merge free intervals starting from the parent up.
	t.mergeFreeStartingAt(parent)

	return true
}
*/

// tryAbsorb checks if the parent can absorb the Interval of a deleted child node.
func (t *Tree) tryAbsorb(node *Tree) {
	// This function will implement the logic of absorbing intervals if applicable,
	// modifying the parent node's interval or merging intervals from its children if necessary.

	// Placeholder implementation.
}

// mergeFreeStartingAt attempts to merge free intervals starting from the given node upwards.
func (t *Tree) mergeFreeStartingAt(node *Tree) {
	// Implement the logic of merging free intervals, working up the tree to ensure that
	// the free intervals are correctly merged after a deletion.

	// This could involve checking siblings and potentially merging them into the parent,
	// then continuing upwards.

	// Placeholder implementation.
}

// Additional helper functions might be necessary to appropriately manage the tree's intervals
// and structure, especially considering specific criteria not fully detailed here.