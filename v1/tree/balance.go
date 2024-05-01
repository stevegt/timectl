package tree

import (
	"math"

	. "github.com/stevegt/goadapt"
)

// rebalance performs the DSW (Day/Stout/Warren) algorithm to rebalance the tree.
func (t *Tree) rebalance() (out *Tree) {
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
	floor := math.Floor
	// log2 := math.Log2

	// number of nodes in a balanced tree of height h is:
	// n = 2^h - 1
	// solving for h to get the final height of the tree:
	// h = log2(n + 1)
	// targetHeight := log2(sizef + 1)

	// We rotate every other node to the left to build the tree, so
	// each compression (round of rotations) will reduce the height of
	// the tree by half.  Looked at another way, we'll need to do
	// m = n/2 rotations to reduce the height of the tree in the first
	// compression, then m/2 rotations in the next compression, and so
	// on.
	rotations := int(floor(sizef / 2.0))

	for ; rotations > 1; rotations /= 2 {
		out = out.compress(rotations)
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

func (t *Tree) compress(rotations int) (out *Tree) {

	if rotations == 0 || t == nil || t.Right == nil {
		return t
	}

	Pf("compress: rotations %d\n", int(rotations))

	// new root is the current root's right child
	out = t.Right

	// we're going to rotate the odd nodes to the left, so we need to
	// keep track of the previous even node so we can attach the next
	// even node to it.
	var prevEven *Tree

	A := t
	// do the rotations
	for i := 0; i < rotations; i++ {
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
			prevEven.Right = B
		}
		prevEven = B

		// C becomes the new A
		A = C
	}

	// ShowDot(out, false)

	return
}
