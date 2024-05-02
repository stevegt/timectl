package tree

import (
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// Delete removes an interval from the tree and returns true if the interval was successfully removed.
func (t *Tree) Delete(interval interval.Interval) (err error) {
	defer Return(&err)
	t.Mu.Lock()
	defer t.Mu.Unlock()

	path, found := t.findExact(interval, nil)
	Assert(found != nil, "Interval not found: %v", interval)

	// Free the node.
	err = t.free(found)
	Assert(err == nil, "Error freeing interval: %v", err)

	// merge with free siblings
	var parent *Tree
	if len(path) > 0 {
		parent = path[len(path)-1]
	} else {
		parent = t
	}
	parent.mergeFree()

	return nil
}

func (t *Tree) free(node *Tree) error {
	freeInterval := interval.NewInterval(node.Start(), node.End(), 0)
	node.Interval = freeInterval
	node.setMinMax()
	return nil
}
