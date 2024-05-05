package tree

import (
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// Tree is an interval tree that stores intervals and allows for fast
// lookup of intervals given an interval.  All inserted intervals are
// stored in leaf nodes; internal nodes store intervals that span the
// intervals of their children.  This allows for fast lookup of
// intervals that overlap a given interval.  Upon insertion of an
// interval, an existing leaf node is split into two leaf nodes, the
// existing interval is moved to one of the two leaf nodes, and the
// new interval is inserted into the other leaf node.  Left node
// intervals always have start times less than or equal to the right
// node interval start times.  If the start times are equal, the left
// node interval end time is less than or equal to the right node
// interval end time.

// test contiguous as a standalone function without a tree
func TestContiguous(t *testing.T) {
	// create some intervals
	ivs := []interval.IInterval{
		NewInterval("2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 0),
		NewInterval("2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1),
		NewInterval("2024-01-01T11:00:00Z", "2024-01-01T12:00:00Z", 0),
		NewInterval("2024-01-01T12:00:00Z", "2024-01-01T15:00:00Z", 0),
		NewInterval("2024-01-01T15:00:00Z", "2024-01-01T16:00:00Z", 1),
	}

	// put them in tree nodes
	nodes := make([]*Node, len(ivs))
	for i, iv := range ivs {
		nodes[i] = newTreeFromInterval(iv)
	}
	Tassert(t, len(nodes) == 5, "Expected 5 nodes, got %d", len(nodes))

	// keep only the free intervals
	nodeChan := slice2chan(nodes)
	freeChan := filter(nodeChan, func(node *Node) bool {
		iv := node.Interval()
		return iv.Priority() < 1
	})
	freeNodes := chan2slice(freeChan)
	Tassert(t, len(freeNodes) == 3, "Expected 3 free nodes, got %d", len(freeNodes))

	// find contiguous nodes of at least 30 minutes
	freeChan = slice2chan(freeNodes)
	free30 := chan2slice(contiguous(freeChan, 30*time.Minute))
	Tassert(t, len(free30) == 1, "Expected 1 free30 node, got %d", len(free30))

	// find contiguous nodes of at least 60 minutes
	freeChan = slice2chan(freeNodes)
	free60 := chan2slice(contiguous(freeChan, 60*time.Minute))
	Tassert(t, len(free60) == 1, "Expected 1 free60 node, got %d", len(free60))

	// find contiguous nodes of at least 90 minutes
	freeChan = slice2chan(freeNodes)
	free90 := chan2slice(contiguous(freeChan, 90*time.Minute))
	Tassert(t, len(free90) == 2, "Expected 2 free90 nodes, got %d", len(free90))

}

// test accumulator
func TestAccumulator(t *testing.T) {
	top := NewTree()

	// accumulate collects nodes in the tree that overlap the given
	// range.  The nodes are collected in order of start time.

	Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:15:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T10:15:00Z")
	Ck(err)

	// get the nodes that overlap the range
	c1 := top.accumulate(true, searchStart, searchEnd)
	nodes := chan2slice(c1)

	// check that we got the right number of nodes
	Tassert(t, len(nodes) == 3, "Expected 3 nodes, got %d", len(nodes))

}

// test filter
func TestFilter(t *testing.T) {
	top := NewTree()

	// filter returns a channel of nodes from the input channel
	// that pass the filter function.

	Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 1)
	Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 2)

	fn := func(t *Node) bool {
		iv := t.Interval()
		return iv.Priority() < 2
	}

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:15:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T10:15:00Z")
	Ck(err)

	c1 := top.accumulate(true, searchStart, searchEnd)
	c2 := filter(c1, fn)
	i2 := chan2slice(c2)

	// check that we got the right number of nodes
	Tassert(t, len(i2) == 2, "Expected 2 nodes, got %d", len(i2))

}

// test contiguous filter
func TestContiguousFilter(t *testing.T) {
	top := NewTree()

	Insert(top, "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", 1)
	Insert(top, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", 2)
	Insert(top, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Insert(top, "2024-01-01T12:15:00Z", "2024-01-01T13:00:00Z", 1)

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:45:00Z")
	Ck(err)

	// get the nodes that overlap the range
	acc := top.accumulate(true, searchStart, searchEnd)
	// filter the nodes to only include those with a priority less than 2
	low := filter(acc, func(t *Node) bool {
		iv := t.Interval()
		return iv.Priority() < 2
	})
	// filter the nodes to only include those that are contiguous
	// for at least N minutes
	cont := contiguous(low, 120*time.Minute)
	res := chan2slice(cont)

	// check that we got the right number of nodes
	Tassert(t, len(res) == 4, "Expected 4 nodes, got %d", len(res))

	// check that we got the right nodes
	var ivs []interval.IInterval
	for _, n := range res {
		ivs = append(ivs, n.Interval())
	}
	err = Match(ivs[0], "2024-01-01T11:00:00Z", "2024-01-01T11:30:00Z", 0)
	Tassert(t, err == nil, err)
	err = Match(ivs[1], "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", 1)
	Tassert(t, err == nil, err)
	err = Match(ivs[2], "2024-01-01T12:00:00Z", "2024-01-01T12:15:00Z", 0)
	Tassert(t, err == nil, err)
	err = Match(ivs[3], "2024-01-01T12:15:00Z", "2024-01-01T13:00:00Z", 1)
	Tassert(t, err == nil, err)
}
