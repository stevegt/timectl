package tree

// rebalance performs the DSW (Day/Stout/Warren) algorithm to rebalance the tree.
func (t *Tree) rebalance() {
	t.treeToVine()
	t.vineToTree()
}

// treeToVine converts the tree into a "vine" (a sorted linked list) using right rotations.
func (t *Tree) treeToVine() (out *Tree) {
	if t == nil {
		return
	}
	out = t
	for out.Left != nil {
		out = out.rotateRight()
	}
	out.Right = out.Right.treeToVine()
	return
}

// vineToTree converts the "vine" back into a balanced tree using left rotations.
func (t *Tree) vineToTree() {
	n := 0
	for temp := t; temp != nil; temp = temp.Right {
		n++
	}
	leafNodes := n + 1 - (1 << (log2(n + 1))) // Number of "incomplete" nodes at the bottom level of the tree
	t.compress(leafNodes)

	// Now, double the tree size until it's just less than 2^k - 1
	for n > 1 { // Whilst there are still nodes to be rotated into the tree
		n >>= 1 // Divide by 2 using a bitwise right shift
		t.compress(n)
	}
}

// compress performs left rotations to construct the tree from the "vine".
func (t *Tree) compress(count int) {
	scanner := t
	for i := 0; i < count; i++ {
		child := scanner.Right
		scanner.Right = child.Right
		scanner = scanner.Right
		child.Right = scanner.Left
		scanner.Left = child
	}
}

// log2 computes the binary logarithm of n using bit shifting.
func log2(n int) int {
	result := 0
	for n > 1 {
		n >>= 1
		result++
	}
	return result
}
