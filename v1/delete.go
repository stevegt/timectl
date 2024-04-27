package timectl

// Delete removes an interval from the tree and returns true if the interval was successfully removed.
func (t *Tree) Delete(interval Interval) (ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// find the interval to delete
	path, found := t.findExact(interval, nil)
	if found == nil {
		return false
	}
	return t.delete(path, found)
}

// delete removes a node from the tree and returns true if the node was successfully removed.
func (t *Tree) delete(path []*Tree, node *Tree) (ok bool) {
	return
}

func (t *Tree) rm(path []*Tree, node *Tree) (ok bool) {
	return
}
