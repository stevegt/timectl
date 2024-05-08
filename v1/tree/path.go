package tree

import (
	. "github.com/stevegt/goadapt"
)

// Path is a slice of Nodes originating from the root of a Tree.
type Path []*Node

// Clone returns a copy of the path.
func (p Path) Clone() Path {
	newPath := make(Path, len(p))
	copy(newPath, p)
	return newPath
}

// Append returns a new Path with the given nodes appended to the end.
func (p Path) Append(nodes ...*Node) Path {
	// because append may reallocate the underlying array, we need to
	// use copy instead of append to avoid modifying the original path
	newPath := make(Path, len(p)+len(nodes))
	copy(newPath, p)
	for i, n := range nodes {
		newPath[len(p)+i] = n
	}
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
// tree has only a root node, the path is {T}.  If the tree has a
// root node and a left child, the path is {T, L}, etc.
func (p Path) Nav() (nav []string) {
	nav = make([]string, len(p))
	var parent *Node
	for i, node := range p {
		switch {
		case parent == nil:
			nav[i] = "t"
		case node == parent.Left():
			nav[i] = "l"
		case node == parent.Right():
			nav[i] = "r"
		default:
			Assert(false, "Node %v is not a child of %v", node, parent)
		}
		parent = node
	}
	return nav
}
