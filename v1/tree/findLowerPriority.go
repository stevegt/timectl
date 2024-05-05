package tree

import (
	"time"
)

// FindLowerPriority returns a contiguous set of nodes that have a
// lower priority than the given priority.  The start time of the
// first node is on or before minStart, and the end time of the last
// node is on or after maxEnd.  The nodes must total at least the
// given duration, and may be longer.  If first is true, then the
// search starts at minStart and proceeds in order, otherwise the
// search starts at maxEnd and proceeds in reverse order.
// XXX this should be refactored to find and return a tree instead of
// a slice; the common parent of the set will always be a member of
// the set.
func (t *Node) FindLowerPriority(first bool, searchStart, searchEnd time.Time, duration time.Duration, priority float64) []*Node {
	t.mu.Lock()
	defer t.mu.Unlock()

	// if the search range fits entirely within the left or right
	// child, then recurse into that child.
	for _, child := range []*Node{t.left, t.right} {
		if child == nil {
			continue
		}
		if searchStart.Before(child.minStart) {
			// child starts too late
			continue
		}
		if searchEnd.After(child.MaxEnd()) {
			// child ends too soon
			continue
		}
		// search range fits entirely within child
		return child.FindLowerPriority(first, searchStart, searchEnd, duration, priority)
	}

	// manage a sliding window of a candidate node set
	var window []*Node
	var sum time.Duration
	// iterate over the nodes in the tree; either in order or reverse
	// order depending on the value of first.
	iter := NewIterator(t, first)
	for {
		node := iter.Next()
		if node == nil {
			break
		}
		// if the node priority is too high, then reset the window
		if node.Interval().Priority() >= priority {
			window = nil
			sum = 0
			continue
		}
		// get the overlap of the node with the search range
		overlap := node.Interval().OverlapDuration(searchStart, searchEnd)
		// if the node does not overlap the search range, then reset
		// the window
		if overlap == 0 {
			window = nil
			sum = 0
			continue
		}
		// add the node to the window
		window = append(window, node)
		sum += overlap
		// if the window is long enough, then return the window
		if sum >= duration {
			if !first {
				// reverse the window if we are iterating in reverse order
				// so that the nodes are in order of start time.
				newWindow := make([]*Node, len(window))
				for i, j := len(window)-1, 0; i >= 0; i, j = i-1, j+1 {
					newWindow[j] = window[i]
				}
				window = newWindow
			}
			return window
		}
	}
	return nil
}

// Iterator allows iterating over the nodes in the tree in either
// forward or reverse order.  If fwd is true, then the iterator will
// iterate in forward order, otherwise it will iterate in reverse
// order.
type Iterator struct {
	Tree *Node
	path []*Node
	Fwd  bool
}

// NewIterator returns a new iterator for the given tree and direction.
func NewIterator(t *Node, fwd bool) *Iterator {
	it := &Iterator{Tree: t, Fwd: fwd}
	// find the path to the first or last node in the tree
	it.path = t.buildpath(fwd)
	return it
}

// buildpath builds a path to the first or last node in the tree.
func (t *Node) buildpath(fwd bool) []*Node {
	node := t
	path := []*Node{node}
	if fwd {
		for node.left != nil {
			node = node.left
			path = append(path, node)
		}
	} else {
		for node.right != nil {
			node = node.right
			path = append(path, node)
		}
	}
	return path
}

// Next returns the next node in the tree.  If the iterator is in
// forward mode, then the nodes are returned in order of start time.
// If the iterator is in reverse mode, then the nodes are returned in
// reverse order of start time.
func (it *Iterator) Next() *Node {
	if len(it.path) == 0 {
		return nil
	}
	res := it.path[len(it.path)-1]
	if it.Fwd {
		if res.right != nil {
			it.path = append(it.path, res.right.buildpath(it.Fwd)...)
		} else {
			// pop nodes off the tail of the path until we find a node
			// that starts later than res
			for {
				try := it.path[len(it.path)-1]
				if try.Interval().Start().After(res.Interval().Start()) {
					break
				}
				it.path = it.path[:len(it.path)-1]
				if len(it.path) == 0 {
					break
				}
			}
		}
	} else {
		if res.left != nil {
			it.path = append(it.path, res.left.buildpath(it.Fwd)...)
		} else {
			// pop nodes off the tail of the path until we find a node
			// that starts earlier than res
			for {
				try := it.path[len(it.path)-1]
				if try.Interval().Start().Before(res.Interval().Start()) {
					break
				}
				it.path = it.path[:len(it.path)-1]
				if len(it.path) == 0 {
					break
				}
			}
		}
	}
	return res
}

func (t *Node) XXXFindLowerPriority(first bool, searchStart, searchEnd time.Time, duration time.Duration, priority float64) []*Node {
	t.mu.Lock()
	defer t.mu.Unlock()

	// get the nodes that overlap the range
	acc := t.accumulate(first, searchStart, searchEnd)

	// filter the nodes to only include those with a priority less
	// than priority
	low := filter(acc, func(node *Node) bool {
		return node.Interval().Priority() < priority
	})

	// filter the nodes to only include those that are contiguous
	// for at least duration
	cont := contiguous(low, duration)
	if !first {
		cont = reverse(cont)
	}

	res := chan2slice(cont)
	return res
}

// accumulate returns a channel of nodes in the tree that wrap the
// given range of start and end times. The nodes are returned in order
// of start time.
func (t *Node) accumulate(fwd bool, start, end time.Time) (out <-chan *Node) {

	// filter function to check if an interval overlaps the given range.
	filterFn := func(t *Node) bool {
		i := t.Interval()
		return i.OverlapsRange(start, end)
	}

	allNodes := t.allNodes(fwd, start, end)
	out = filter(allNodes, filterFn)
	return
}

// slice2chan converts a slice of nodes to a channel of nodes.
func slice2chan(nodes []*Node) <-chan *Node {
	ch := make(chan *Node)
	go func() {
		for _, i := range nodes {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

// chan2slice converts a channel of nodes to a slice of nodes.
func chan2slice(ch <-chan *Node) []*Node {
	nodes := []*Node{}
	for i := range ch {
		nodes = append(nodes, i)
	}
	return nodes
}

// filter filters nodes from a channel of nodes based on a
// filter function and returns a channel of nodes.
func filter(in <-chan *Node, filterFn func(tree *Node) bool) <-chan *Node {
	out := make(chan *Node)
	go func() {
		for i := range in {
			if filterFn(i) {
				out <- i
			}
		}
		close(out)
	}()
	return out
}

// contiguous returns a channel of nodes from the input
// channel that are contiguous and have a total duration of at
// least the given duration.  The nodes may be provided in
// either forward or reverse order, and will be returned in the
// order they are provided.  The channel is closed when the
// first matching set of nodes is found.
func contiguous(ch <-chan *Node, duration time.Duration) <-chan *Node {
	out := make(chan *Node)
	go func() {
		defer close(out)
		var sum time.Duration
		var nodes []*Node
		for n := range ch {
			if sum >= duration {
				break
			}
			i := n.Interval()
			if len(nodes) == 0 {
				nodes = append(nodes, n)
				sum = i.Duration()
				continue
			}
			lastInterval := nodes[len(nodes)-1].Interval()
			nStart := n.Interval().Start()
			nEnd := n.Interval().End()
			okFwd := nStart.Equal(lastInterval.End())
			okRev := nEnd.Equal(lastInterval.Start())
			if okFwd || okRev {
				nodes = append(nodes, n)
				sum += i.Duration()
			} else {
				nodes = []*Node{n}
				sum = i.Duration()
			}
		}
		if sum >= duration {
			for _, n := range nodes {
				out <- n
			}
		}
	}()
	return out
}

// reverse reverses the order of nodes in a channel of nodes.
func reverse(ch <-chan *Node) <-chan *Node {
	out := make(chan *Node)
	go func() {
		defer close(out)
		nodes := []*Node{}
		for i := range ch {
			nodes = append(nodes, i)
		}
		for i := len(nodes) - 1; i >= 0; i-- {
			out <- nodes[i]
		}
	}()
	return out
}
