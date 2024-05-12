package tree

import (
	"time"
	// . "github.com/stevegt/goadapt"
)

// FindLowerPriority returns a contiguous set of nodes that have a
// lower priority than the given priority.  The start time of the
// first node is on or before minStart, and the end time of the last
// node is on or after maxEnd.  The nodes must total at least the
// given duration, and may be longer.  If first is true, then the
// search starts at minStart and proceeds in order, otherwise the
// search starts at maxEnd and proceeds in reverse order.
// XXX this should be refactored to find and return a path to a
// subtree instead of a slice; the common parent of the set will
// always be a member of the set.
func (t *Node) FindLowerPriority(first bool, searchStart, searchEnd time.Time, duration time.Duration, priority float64) (window []*Node, out Path) {
	t.mu.Lock()
	defer t.mu.Unlock()

	out = t.path

	// if the search range fits entirely within the left or right
	// child, then recurse into that child.
	for _, child := range []*Node{t.Left(), t.Right()} {
		if child == nil {
			continue
		}
		if searchStart.Before(child.MinStart()) {
			// child starts too late
			continue
		}
		if searchEnd.After(child.MaxEnd()) {
			// child ends too soon
			continue
		}
		// search range fits entirely within child
		var subPath Path
		window, subPath = child.FindLowerPriority(first, searchStart, searchEnd, duration, priority)
		out = out.Append(subPath...)
		return
	}

	// manage a sliding window of a candidate node set
	var sum time.Duration
	// iterate over the nodes in the tree; either in order or reverse
	// order depending on the value of first.
	iter := NewIterator(t, first)
	windowFound := false
	for {
		path := iter.Next()
		node := path.Last()
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
			windowFound = true
			break
		}
	}
	if !windowFound {
		return nil, out
	}
	return window, out

	// XXX temporary: keep the window stuff above, but now redo the
	// search by ignoring the window and building a subtree instead.

	/*
		// If we get here, then we've found the node whose subtree
		// best fits the search range.  Now we need to trim the subtree,
		// removing nodes that don't fit in the search range.
		// trim first
		for {
			firstNode := out.FirstNode()
			if firstNode == nil {
				return nil, nil
			}
			if searchStart.Before(firstNode.End()) {
				break
			}
			var err error
			out, err = out.deleteNode(firstNode)
			Assert(err == nil, err)
		}
		// trim last
		for {
			lastNode := out.LastNode()
			if lastNode == nil {
				return nil, nil
			}
			if searchEnd.After(lastNode.Start()) {
				break
			}
			var err error
			out, err = out.deleteNode(lastNode)
			Assert(err == nil, err)
		}

		// Now we need to find the node in the subtree with a maxPriority
		// less than the given priority and maxEnd - minStart >= duration.

		subtreeDuration := out.MaxEnd().Sub(out.MinStart())
		if subtreeDuration < duration {
			// subtree duration is too short; no need to try children
			// because they will be shorter
			return nil, nil
		}

		// get a list of children sorted according to the value of first
		children := []*Node{out.Left(), out.Right()}
		if !first {
			children = []*Node{out.Right(), out.Left()}
		}

		if out.MaxPriority() >= priority {
			// priority is too high; try children
			for _, child := range children {
				if child == nil {
					continue
				}
				if child.MaxPriority() < priority {
					// child has a lower priority; recurse into it
					// return child.FindLowerPriority(first, searchStart, searchEnd, duration, priority)
					// XXX temporary:  build a slice of nodes from the subtree
					_, out = child.FindLowerPriority(first, searchStart, searchEnd, duration, priority)
				}
			}
			return nil, nil
		}

		// got it!

		// XXX temporary:  build a slice of nodes from the subtree
		iter := NewIterator(out, first)
		for {
			path := iter.Next()
			node := path.Last()
			if node == nil {
				break
			}
			// add the node to the window
			window = append(window, node)
		}

		return
	*/
}

// Iterator allows iterating over the nodes in the tree in either
// forward or reverse order.  If fwd is true, then the iterator will
// iterate in forward order, otherwise it will iterate in reverse
// order.
type Iterator struct {
	Tree *Node
	path Path
	Fwd  bool
}

// NewIterator returns a new iterator for the given tree and direction.
func NewIterator(t *Node, fwd bool) *Iterator {
	it := &Iterator{Tree: t, Fwd: fwd}
	// find the path to the first or last node in the tree
	it.path = t.buildpath(fwd)
	return it
}

// buildpath builds a path to the first or last node in the subtree
// rooted at the given node.  If left is true, then the path is built
// to the leftmost node in the subtree, otherwise the path is built to
// the rightmost node in the subtree.
func (t *Node) buildpath(left bool) Path {
	node := t
	path := Path{node}
	if left {
		for node.Left() != nil {
			node = node.Left()
			path = path.Append(node)
		}
	} else {
		for node.Right() != nil {
			node = node.Right()
			path = path.Append(node)
		}
	}
	return path
}

// Next returns the path to the next node in the tree.  If the
// iterator is in forward mode, then the nodes are returned in order
// of start time. If the iterator is in reverse mode, then the nodes
// are returned in reverse order of start time.
func (it *Iterator) Next() (path Path) {
	if len(it.path) == 0 {
		return nil
	}

	// return the last node in the path
	node := it.path.Last()
	path = it.path.Clone()

	// configure the path for the next iteration
	if it.Fwd {
		if node.Right() != nil {
			// node has a right child; get the path to the right
			// child's leftmost node and append it to the path for
			// next time
			leftPath := node.Right().buildpath(it.Fwd)
			it.path = it.path.Append(leftPath...)
		} else {
			// pop nodes off the tail of the path until we find a node
			// that starts later than res
			for {
				try := it.path.Last()
				if try.Interval().Start().After(node.Interval().Start()) {
					break
				}
				it.path = it.path.Pop()
				if len(it.path) == 0 {
					break
				}
			}
		}
	} else {
		if node.Left() != nil {
			rightPath := node.Left().buildpath(it.Fwd)
			it.path = it.path.Append(rightPath...)
		} else {
			// pop nodes off the tail of the path until we find a node
			// that starts earlier than res
			for {
				try := it.path.Last()
				if try.Interval().Start().Before(node.Interval().Start()) {
					break
				}
				it.path = it.path.Pop()
				if len(it.path) == 0 {
					break
				}
			}
		}
	}
	return
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