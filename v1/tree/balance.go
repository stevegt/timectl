package tree

import (
	"math"
	// . "github.com/stevegt/goadapt"
)

// rebalance uses the height and size fields to balance the tree.
func (t *Node) rebalance() (out *Node) {
	if t == nil {
		return
	}

	out = t

	for i := 0; i < 200; i++ {
		var leftSize, rightSize int
		var leftHeight, rightHeight int
		if out.left != nil {
			out.SetLeft(out.left.rebalance())
			leftSize = out.left.Size()
			leftHeight = out.left.height
		}
		if out.right != nil {
			out.SetRight(out.right.rebalance())
			rightSize = out.right.Size()
			rightHeight = out.right.height
		}
		// Pf("rebalance: %d - %d\n", leftHeight, rightHeight)
		if leftHeight-rightHeight > 1 {
			// Pf("rotateRight leftHeight: %d - rightHeight: %d\n", leftHeight, rightHeight)
			out = out.rotateRight()
			continue
		}
		if rightHeight-leftHeight > 1 {
			// Pf("rotateLeft rightHeight: %d - leftHeight: %d\n", rightHeight, leftHeight)
			out = out.rotateLeft()
			continue
		}
		if false && leftSize-rightSize > 1 {
			// Pf("rotateRight leftSize: %d - rightSize: %d\n", leftSize, rightSize)
			out = out.rotateRight()
			continue
		}
		if false && rightSize-leftSize > 1 {
			// Pf("rotateLeft rightSize: %d - leftSize: %d\n", rightSize, leftSize)
			out = out.rotateLeft()
			continue
		}
		break
	}
	return
}

// rebalanceDSW performs the DSW (Day/Stout/Warren) algorithm to rebalance the tree.
func (t *Node) rebalanceDSW() (out *Node) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var size int
	out, size = t.treeToVine()
	// ShowDot(out, false)
	out = out.vineToTree(size)
	return
}

// treeToVine converts the tree into a "vine" (a sorted linked list) using right rotations.
func (t *Node) treeToVine() (out *Node, size int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t == nil {
		return
	}
	out = t
	// rotate the left children to the right
	for out.left != nil {
		out = out.rotateRight()
	}
	// continue down the right side of the tree
	if out.right != nil {
		var right *Node
		right, size = out.right.treeToVine()
		out.SetRight(right)
	}
	size++
	return
}

// vineToTree converts the "vine" back into a balanced tree using left rotations.
func (t *Node) vineToTree(size int) (out *Node) {
	out = t
	sizef := float64(size)
	// pow := math.Pow
	// floor := math.Floor
	log2 := math.Log2

	// number of nodes in a balanced tree of height h is:
	// n = 2^h - 1
	// solving for h to get the final height of the tree:
	// h = log2(n + 1)
	targetHeight := int(log2(sizef + 1))

	// We rotate every other node to the left to build the tree, so
	// each compression (round of rotations) will reduce the height of
	// the tree by half.  Looked at another way, we'll need to do
	// m = n/2 rotations to reduce the height of the tree in the first
	// compression, then m/2 rotations in the next compression, and so
	// on.
	// rotations := int(floor(sizef / 2.0))
	// for ; rotations > 1; rotations /= 2 {
	// 	  out = out.compress(rotations)
	// }

	// Hang on.  The only reason we're doing all this math is so we
	// can use it in O() analysis.  We don't need to do that.  We can
	// just keep compressing the tree until we're done.  Geez.
	for done := false; !done; {
		out, done = out.compress(targetHeight)
	}

	// One last check to make sure the tree is balanced.
	if out.right != nil && out.left != nil {
		for out.right.CalcHeight() > out.left.CalcHeight() {
			out = out.rotateLeft()
		}
	}
	return
}

// compress performs left rotations using rotateLeft on the odd nodes
// to compress the tree.
//
// we start like this:
// n = 7
//
//     A
//      \
//       B
//        \
//         C
//          \
//           D
//            \
//             E
//              \
//	    		 F
//                \
//                 G
//
// first compression: m = 7 - 2^floor(log2(7)) = 7 - 4 = 3 rotations
//
//     B
//    / \
//   A   D
//      / \
//     C   F
//        / \
//       E   G
//
// second compression: m = 3 - 2^floor(log2(3)) = 3 - 2 = 1 rotation
//
//     D
//    / \
//   B   F
//  / \ / \
// A  C E  G
//

func (t *Node) compress(targetHeight int) (out *Node, done bool) {

	if t == nil || t.right == nil {
		return t, true
	}

	// new root is the current root's right child
	out = t.right

	// we're going to rotate the odd nodes to the left, so we need to
	// keep track of the previous even node so we can attach the next
	// even node to it.
	var prevEven *Node

	A := t
	// do the rotations
	height := 0
	for {
		// Odd node, e.g. (A): rotate the node, promoting
		// and returning (B), which is even.  We'll need to
		// hang onto (B) so we can attach the next even node
		// to it.
		//
		//           B
		//          / \
		//   (odd) A   C
		//              \
		//               D
		//
		if A == nil || A.right == nil {
			break
		}
		B := A.rotateLeft()
		C := B.right

		// attach B to the previous even node (if there is one)
		if prevEven != nil {
			prevEven.SetRight(B)
		}
		prevEven = B

		// C becomes the new A
		A = C

		// increment the height by 2 since we're rotating the odd nodes
		height += 2
	}
	done = height <= targetHeight

	// ShowDot(out, false)

	return
}

// absInt returns the absolute value of an integer.
func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
