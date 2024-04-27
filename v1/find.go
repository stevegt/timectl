package timectl

// . "github.com/stevegt/goadapt"

// FindExact returns the tree node containing the exact interval that
// matches the given interval, along with the path of ancestor nodes.
// If the exact interval is not found, then the found node is nil and
// the path node ends with the node where the interval would be
// inserted.  If the exact interval is in the root node, then the path
// is nil.  If the tree is empty, then both are nil.
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

	// if the interval starts before the parent's synthetic interval ends, then
	// return the parent node as the place where the interval would
	// be inserted
	if len(pathIn) > 0 {
		parent := pathIn[len(pathIn)-1]
		if interval.Start().Before(parent.treeEnd()) {
			return pathIn, nil
		}
	}

	return nil, nil
}
