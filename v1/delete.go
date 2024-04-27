package timectl

// . "github.com/stevegt/goadapt"

// FindExact returns the tree node containing the exact interval
// that matches the given interval, along with the parent node.
// If the exact interval is not found, then the found node is nil
// and the parent node is the node where the interval would be
// inserted.  If the exact interval is in the root node, then the
// parent node is nil.  If the tree is empty, then both nodes are nil.
func (t *Tree) FindExact(interval Interval) (found, parent *Tree) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.findExact(interval, nil)
}

// findExact is a non-threadsafe version of FindExact for internal
// use.  The parent parameter is used to track the parent of the
// current node during recursion.
func (t *Tree) findExact(interval Interval, parentIn *Tree) (found, parent *Tree) {

	if t.leafInterval == nil {
		// non-leaf node
		// try left
		if t.left != nil {
			found, parent = t.left.findExact(interval, t)
		}
		// try right
		if found == nil && t.right != nil {
			found, parent = t.right.findExact(interval, t)
		}
		return
	}

	// leaf node
	if t.leafInterval.Equal(interval) {
		return t, parentIn
	}

	// if the interval starts before the parent's synthetic interval ends, then
	// return the parent node as the place where the interval would
	// be inserted
	if parentIn != nil && interval.Start().Before(parentIn.End()) {
		return nil, parentIn
	}

	return nil, nil
}

func (t *Tree) Delete(interval Interval) (ok bool) {
	return
}

func (t *Tree) delete(interval Interval) (ok bool) {
	return
}
