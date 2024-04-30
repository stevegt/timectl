package tree

import (
	"math"

	. "github.com/stevegt/goadapt"
)

// rebalance performs the DSW (Day/Stout/Warren) algorithm to rebalance the tree.
func (t *Tree) rebalance() (out *Tree) {
	var size int
	out, size = t.treeToVine()
	out = t.vineToTree(size)
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
	pow := math.Pow
	floor := math.Floor
	log2 := math.Log2
	rotations := sizef + 1 - pow(floor(log2(sizef+1)), 2)
	// out = out.compress(rotations)
	rotations = sizef - rotations
	Pf("size: %d, rotations: %f\n", size, rotations)
	for rotations > 1 {
		rotations = rotations / 2
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

func (t *Tree) compress(rotations float64) (out *Tree) {

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
	for i := 0; i < int(rotations); i++ {
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

	return
}
