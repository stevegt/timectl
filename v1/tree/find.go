package timectl

import (
	"github.com/stevegt/timectl/interval"
)

// . "github.com/stevegt/goadapt"

// FindExact returns the tree node containing the exact interval that
// matches the given interval, along with the path of ancestor nodes.
// If the exact interval is not found, then the path and found node
// are both nil.  If the exact interval is in the root node, then the
// Muth is nil.
func (t *Tree) FindExact(interval interval.Interval) (path []*Tree, found *Tree) {
	t.Mu.RLock()
	defer t.Mu.RUnlock()
	return t.findExact(interval, nil)
}

// findExact is a non-threadsafe version of FindExact for internal
// use.  The path parameter is used to track the path to the
// current node during recursion.
func (t *Tree) findExact(interval interval.Interval, pathIn []*Tree) (pathOut []*Tree, found *Tree) {

	if t.Interval == nil {
		// non-leaf node
		// try left
		path := append(pathIn, t)
		if t.Left != nil {
			pathOut, found = t.Left.findExact(interval, path)
		}
		// try right
		if found == nil && t.Right != nil {
			pathOut, found = t.Right.findExact(interval, path)
		}
		return
	}

	// leaf node
	if t.Interval.Equal(interval) {
		return pathIn, t
	}

	return nil, nil
}
