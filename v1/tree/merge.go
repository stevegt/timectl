package tree

import "github.com/stevegt/timectl/interval"

// mergeFree merges adjacent free intervals in the tree.
func (t *Tree) mergeFree() {

	// Check for nil because this function could be called on a nil receiver due to defer in Tree operations.
	if t == nil {
		return
	}

	// merge free intervals on the left
	if t.Left != nil {
		t.Left.mergeFree()
	}

	// merge free intervals on the right
	if t.Right != nil {
		t.Right.mergeFree()
	}

	// let's say we have this tree:
	//
	//             a
	//            / \
	//           B   C
	//          / \
	//         D   e
	//
	// ...where lower case letters are free intervals and upper case letters are busy intervals.
	//
	// We want to merge the free intervals to get:
	//
	//             a
	//            / \
	//           B   C
	//          /
	//         D
	//
	// We can only merge a with the rightmost leaf of its left child,
	// or a with the leftmost leaf of its right child.
	//

	// try the rightmost leaf of the left child
	parent := t
	leaf := t.Left
	// find the rightmost leaf
	for leaf != nil && leaf.Right != nil {
		parent = leaf
		leaf = leaf.Right
	}
	if leaf != nil && !leaf.Busy() {
		// merge leaf with t
		t.Interval = interval.NewInterval(leaf.Interval.Start(), t.Interval.End(), 0)
		// remove leaf
		if parent == t {
			t.Left = nil
		} else {
			parent.Right = nil
		}
		parent.setMinMax()
	}

	// try the leftmost leaf of the right child
	parent = t
	leaf = t.Right
	// find the leftmost leaf
	for leaf != nil && leaf.Left != nil {
		parent = leaf
		leaf = leaf.Left
	}
	if leaf != nil && !leaf.Busy() {
		// merge leaf with t
		t.Interval = interval.NewInterval(t.Interval.Start(), leaf.Interval.End(), 0)
		// remove leaf
		if parent == t {
			t.Right = nil
		} else {
			parent.Left = nil
		}
		parent.setMinMax()
	}

	t.setMinMax()
}
