package tree

import (
	"github.com/stevegt/timectl/interval"
)

// . "github.com/stevegt/goadapt"

// FindExact returns the tree node containing the exact interval that
// matches the given interval, along with the path of ancestor nodes.
// If the exact interval is not found, then the path and found node
// are both nil.  If the exact interval is in the root node, then the
// path is nil.
func (t *Node) FindExact(interval interval.Interval) (path []*Node, found *Node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.findExact(interval, nil)
}

// findExact is a recursive version of FindExact for internal
// use.  The path parameter is used to track the path to the
// current node during recursion.
func (t *Node) findExact(interval interval.Interval, pathIn []*Node) (pathOut []*Node, found *Node) {

	if t.Interval.Equal(interval) {
		return pathIn, t
	}

	path := append(pathIn, t)

	// try left
	if t.left != nil {
		return t.left.findExact(interval, path)
	}
	// try right
	if found == nil && t.right != nil {
		return t.right.findExact(interval, path)
	}

	return nil, nil
}
