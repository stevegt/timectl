package interval

// Tree represents a node in an interval tree.
type Tree struct {
	left  *Tree // Pointer to the left child
	right *Tree // Pointer to the right child
}

// NewTree creates and returns a new Tree
func NewTree() *Tree {
	return &Tree{}
}
