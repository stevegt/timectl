package interval

// Tree represents a node in an interval tree.
type Tree struct {
	// If this is a leaf node, then interval is the interval stored in
	// this node.  If this is not a leaf node, then interval is an
	// interval that spans the left and right children.
	interval *Interval
	left     *Tree // Pointer to the left child
	right    *Tree // Pointer to the right child
}

// NewTree creates and returns a new Tree node without an interval.
// This change will help in ensuring that our tree starts off empty,
// and the root node will be populated on the first insertion.
func NewTree() *Tree {
	return &Tree{}
}
