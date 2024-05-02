package tree

import (
	"math"
	// . "github.com/stevegt/goadapt"
)

// rebalance uses the Height and Size fields to balance the tree.
func (t *Tree) rebalance() (out *Tree) {
	if t == nil {
		return
	}

	out = t

	for i := 0; i < 200; i++ {
		var leftSize, rightSize int
		var leftHeight, rightHeight int
		if out.Left != nil {
			out.SetLeft(out.Left.rebalance())
			leftSize = out.Left.Size
			leftHeight = out.Left.Height
		}
		if out.Right != nil {
			out.SetRight(out.Right.rebalance())
			rightSize = out.Right.Size
			rightHeight = out.Right.Height
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
func (t *Tree) rebalanceDSW() (out *Tree) {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	var size int
	out, size = t.treeToVine()
	// ShowDot(out, false)
	out = out.vineToTree(size)
	return
}

// treeToVine converts the tree into a "vine" (a sorted linked list) using right rotations.
func (t *Tree) treeToVine() (out *Tree, size int) {
	if t == nil {
		return
	}
	out = t
	// rotate the left children to the right
	for out.Left != nil {
		out = out.rotateRight()
	}
	// continue down the right side of the tree
	out.Right, size = out.Right.treeToVine()
	size++
	return
}

// vineToTree converts the "vine" back into a balanced tree using left rotations.
func (t *Tree) vineToTree(size int) (out *Tree) {
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
	if out.Right != nil && out.Left != nil {
		for out.Right.height() > out.Left.height() {
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

func (t *Tree) compress(targetHeight int) (out *Tree, done bool) {

	if t == nil || t.Right == nil {
		return t, true
	}

	// new root is the current root's right child
	out = t.Right
	// old root's parent is now the new root
	t.Parent = out
	// new root's parent is nil
	out.Parent = nil

	// we're going to rotate the odd nodes to the left, so we need to
	// keep track of the previous even node so we can attach the next
	// even node to it.
	var prevEven *Tree

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
		if A == nil || A.Right == nil {
			break
		}
		B := A.rotateLeft()
		C := B.Right

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
