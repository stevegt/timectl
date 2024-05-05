package tree

// Path is a slice of Tree nodes.
type Path []*Node

// Append returns a new Path with the given node appended to the end.
func (p Path) Append(t *Node) Path {
	// because append may reallocate the underlying array, we need to
	// use copy instead of append to avoid modifying the original path
	newPath := make(Path, len(p)+1)
	copy(newPath, p)
	newPath[len(p)] = t
	return newPath
}

// Last returns the last node in the path.
func (p Path) Last() *Node {
	return p[len(p)-1]
}

// String returns a string representation of the path.
func (p Path) String() string {
	var s string
	var parent *Node
	for _, t := range p {
		if parent != nil {
			if t == parent.left {
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
