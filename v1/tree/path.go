package tree

// Path is a slice of Tree nodes.
type Path []*Node

// Clone returns a copy of the path.
func (p Path) Clone() Path {
	newPath := make(Path, len(p))
	copy(newPath, p)
	return newPath
}

// Append returns a new Path with the given node appended to the end.
func (p Path) Append(t *Node) Path {
	// because append may reallocate the underlying array, we need to
	// use copy instead of append to avoid modifying the original path
	newPath := make(Path, len(p)+1)
	copy(newPath, p)
	newPath[len(p)] = t
	return newPath
}

// First returns the first node in the path.
func (p Path) First() *Node {
	if len(p) == 0 {
		return nil
	}
	return p[0]
}

// Last returns the last node in the path.
func (p Path) Last() *Node {
	if len(p) == 0 {
		return nil
	}
	return p[len(p)-1]
}

// String returns a string representation of the path.
func (p Path) String() string {
	var s string
	var parent *Node
	for _, t := range p {
		if parent != nil {
			if t == parent.Left() {
				s += "l"
			} else {
				s += "r"
			}
		} else {
			s += "t"
		}
		parent = t
	}
	return s
}

// Nav returns the navigation directions to get to a node from the
// root.  If the path is empty, it returns an empty slice.  If the
// tree has only a root node, the path is {'t'}.  If the tree has a
// root node and a left child, the path is {'t', 'l'}, etc.
func (p Path) Nav() (nav []rune) {
	nav = make([]rune, len(p))
	var parent *Node
	for i, node := range p {
		switch {
		case parent == nil:
			nav[i] = 't'
		case node == parent.Left():
			nav[i] = 'l'
		case node == parent.Right():
			nav[i] = 'r'
		default:
			Assert(false, "Node %v is not a child of %v", node, parent)
		}
	}
	return nav
}
