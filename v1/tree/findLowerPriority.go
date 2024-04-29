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
func (t *Tree) FindLowerPriority(first bool, searchStart, searchEnd time.Time, duration time.Duration, priority float64) []*Tree {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	// get the nodes that overlap the range
	acc := t.accumulate(first, searchStart, searchEnd)

	// filter the nodes to only include those with a priority less
	// than priority
	low := filter(acc, func(node *Tree) bool {
		return node.Interval.Priority() < priority
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
func (t *Tree) accumulate(fwd bool, start, end time.Time) (out <-chan *Tree) {

	// filter function to check if an interval overlaps the given range.
	filterFn := func(t *Tree) bool {
		i := t.Interval
		return i.OverlapsRange(start, end)
	}

	allNodes := t.allNodes(fwd, start, end)
	out = filter(allNodes, filterFn)
	return
}

// slice2chan converts a slice of nodes to a channel of nodes.
func slice2chan(nodes []*Tree) <-chan *Tree {
	ch := make(chan *Tree)
	go func() {
		for _, i := range nodes {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

// chan2slice converts a channel of nodes to a slice of nodes.
func chan2slice(ch <-chan *Tree) []*Tree {
	nodes := []*Tree{}
	for i := range ch {
		nodes = append(nodes, i)
	}
	return nodes
}

// filter filters nodes from a channel of nodes based on a
// filter function and returns a channel of nodes.
func filter(in <-chan *Tree, filterFn func(tree *Tree) bool) <-chan *Tree {
	out := make(chan *Tree)
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
func contiguous(ch <-chan *Tree, duration time.Duration) <-chan *Tree {
	out := make(chan *Tree)
	go func() {
		defer close(out)
		var sum time.Duration
		var nodes []*Tree
		for n := range ch {
			if sum >= duration {
				break
			}
			i := n.Interval
			if len(nodes) == 0 {
				nodes = append(nodes, n)
				sum = i.Duration()
				continue
			}
			lastInterval := nodes[len(nodes)-1].Interval
			nStart := n.Interval.Start()
			nEnd := n.Interval.End()
			okFwd := nStart.Equal(lastInterval.End())
			okRev := nEnd.Equal(lastInterval.Start())
			if okFwd || okRev {
				nodes = append(nodes, n)
				sum += i.Duration()
			} else {
				nodes = []*Tree{n}
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
func reverse(ch <-chan *Tree) <-chan *Tree {
	out := make(chan *Tree)
	go func() {
		nodes := []*Tree{}
		for i := range ch {
			nodes = append(nodes, i)
		}
		for i := len(nodes) - 1; i >= 0; i-- {
			out <- nodes[i]
		}
		close(out)
	}()
	return out
}
