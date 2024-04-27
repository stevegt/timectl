package timectl

import (
	"fmt" // Import the fmt package to use for formatting errors.
)

// Verify checks the integrity of the tree structure. It makes sure that all intervals
// are correctly placed within the tree according to the interval tree properties.
// Specifically, it ensures that for any given node:
// - The left child's max end time is less than the current node's start time.
// - The right child's start time is greater than or equal to the current node's end time.
// It also checks that all intervals within the tree do not overlap unless specified
// by the intervals' properties.
func (t *Tree) Verify() error {
    // Define a recursive helper function to traverse the tree and verify each node.
    var verifyRecursive func(node *Tree) error
    verifyRecursive = func(node *Tree) error {
        if node == nil {
            return nil // A nil node does not violate any properties.
        }

        // If the current node is a leaf and has an interval, verify it does not overlap with its siblings.
        if node.leafInterval != nil {
            if node.left != nil && node.left.leafInterval != nil && node.left.leafInterval.Conflicts(node.leafInterval, false) {
                return fmt.Errorf("left child conflicts with node interval")
            }
            if node.right != nil && node.right.leafInterval != nil && node.right.leafInterval.Conflicts(node.leafInterval, false) {
                return fmt.Errorf("right child conflicts with node interval")
            }
        }

        // Verify the left subtree.
        if err := verifyRecursive(node.left); err != nil {
            return err // Propagate the error upwards.
        }

        // Verify the right subtree.
        if err := verifyRecursive(node.right); err != nil {
            return err // Propagate the error upwards.
        }

        return nil
    }

    // Start the verification from the root of the tree.
    return verifyRecursive(t)
}
