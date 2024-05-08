package tree

import (
	"github.com/stevegt/timectl/interval"
)

// . "github.com/stevegt/goadapt"

// FindExact returns the path to the node containing the exact
// interval that matches the given interval. If the exact interval is
// not found, then the path is nil.
func (t *Node) FindExact(interval interval.Interval) (path Path) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.findExact(interval, Path{})
}

// findExact is a recursive version of FindExact for internal
// use.  The path parameter is used to track the path to the
// current node during recursion.
func (t *Node) findExact(interval interval.Interval, pathIn Path) (pathOut Path) {

	pathOut = pathIn.Append(t)

	if t.Interval().Equal(interval) {
		return
	}

	// try left
	if t.Left() != nil {
		p := t.Left().findExact(interval, pathOut)
		if p != nil {
			pathOut = p
			return
		}
	}
	// try right
	if t.Right() != nil {
		p := t.Right().findExact(interval, pathOut)
		if p != nil {
			pathOut = p
			return
		}
	}

	return nil
}
