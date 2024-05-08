package tree

import (
	"github.com/stevegt/timectl/interval"
)

// . "github.com/stevegt/goadapt"

// FindExact returns the tree node containing the exact interval that
// matches the given interval, along with the path to the node,
// including the found node.  If the exact interval is not found, then
// the path and found node are both nil.  If the exact interval is in
// the root node, then the path is nil.
func (t *Node) FindExact(interval interval.Interval) (path Path, found *Node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.findExact(interval, Path{})
}

// findExact is a recursive version of FindExact for internal
// use.  The path parameter is used to track the path to the
// current node during recursion.
func (t *Node) findExact(interval interval.Interval, pathIn Path) (pathOut Path, found *Node) {

	pathOut = pathIn.Append(t)

	if t.Interval().Equal(interval) {
		found = t
		return
	}

	// try left
	if t.Left() != nil {
		p, n := t.Left().findExact(interval, pathOut)
		if n != nil {
			pathOut = p
			found = n
			return
		}
	}
	// try right
	if found == nil && t.Right() != nil {
		p, n := t.Right().findExact(interval, pathOut)
		if n != nil {
			pathOut = p
			found = n
			return
		}
	}

	return nil, nil
}
