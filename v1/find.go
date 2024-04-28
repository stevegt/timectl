package timectl

// . "github.com/stevegt/goadapt"

// FindExact returns the tree node containing the exact interval that
// matches the given interval, along with the path of ancestor nodes.
// If the exact interval is not found, then the path and found node
// are both nil.  If the exact interval is in the root node, then the
// path is nil.
func (t *Tree) FindExact(interval Interval) (path []*Tree, found *Tree) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.findExact(interval, nil)
}

// findExact is a non-threadsafe version of FindExact for internal
// use.  The path parameter is used to track the path to the
// current node during recursion.
func (t *Tree) findExact(interval Interval, pathIn []*Tree) (pathOut []*Tree, found *Tree) {

	if t.interval == nil {
		// non-leaf node
		// try left
		path := append(pathIn, t)
		if t.left != nil {
			pathOut, found = t.left.findExact(interval, path)
		}
		// try right
		if found == nil && t.right != nil {
			pathOut, found = t.right.findExact(interval, path)
		}
		return
	}

	// leaf node
	if t.interval.Equal(interval) {
		return pathIn, t
	}

	return nil, nil
}
