package interval

import (
	"fmt"
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
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

// get is a test helper function that returns the tree node at the
// given path in the tree.  pathStr is the path to the node in the
// tree, where 'l' means to go left and 'r' means to go right.  An
// empty pathStr means to return the root node.
func get(tree *Tree, pathStr string) *Tree {
	path := []rune(pathStr)
	if len(path) == 0 {
		return tree
	}
	switch path[0] {
	case 'l':
		return get(tree.left, string(path[1:]))
	case 'r':
		return get(tree.right, string(path[1:]))
	default:
		Assert(false, "Invalid path %v", pathStr)
	}
	return nil
}

// expect is a test helper function that checks if the given tree
// node's interval has the expected start and end times and event
// content. pathStr is the path to the node in the tree, where 'l'
// means to go left and 'r' means to go right.  An empty pathStr means
// to check the root node.
func expect(tree *Tree, pathStr, startStr, endStr string, event any) error {
	node := get(tree, pathStr)
	if node == nil {
		return fmt.Errorf("no node at path: %v", pathStr)
	}
	nodeInterval := node.Interval()
	if nodeInterval.Event() != event {
		return fmt.Errorf("Expected event=%v, got event=%v", event, nodeInterval.Event())
	}
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	ev := NewInterval(start, end, event)
	if !node.Interval().Equal(ev) {
		return fmt.Errorf("Expected %v, got %v", ev, node.Interval())
	}
	return nil
}

// insertExpect is a test helper function that inserts an interval
// into the tree and checks if the tree has the expected structure.
func insertExpect(tree *Tree, pathStr, startStr, endStr string, event any) error {
	interval := insert(tree, startStr, endStr, event)
	if interval == nil {
		return fmt.Errorf("Failed to insert interval")
	}
	return expect(tree, pathStr, startStr, endStr, event)
}

// insert is a test helper function that inserts an interval into the
// tree and returns the interval that was inserted.
func insert(tree *Tree, startStr, endStr string, event any) *Interval {
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	interval := NewInterval(start, end, event)
	// Insert adds a new interval to the tree, adjusting the structure as
	// necessary.  Insertion fails if the new interval conflicts with any
	// existing interval in the tree.
	ok := tree.Insert(interval)
	if !ok {
		return nil
	}
	return interval
}

// Test a tree node with children.
// A tree is a tree of busy and free intervals that
// span the entire range from treeStart to treeEnd.
func TestTreeStructure(t *testing.T) {

	tree := NewTree()
	// insert interval into empty tree -- this should become the left
	// child of the right child of the root node
	err := insertExpect(tree, "rl", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	Tassert(t, err == nil, err)
	// the other nodes should be non-busy
	err = expect(tree, "l", TreeStartStr, "2024-01-01T10:00:00Z", nil)
	Tassert(t, err == nil, err)
	err = expect(tree, "rr", "2024-01-01T11:00:00Z", TreeEndStr, nil)

	/*
		// insert a right interval -- this should cause the new interval
		// to be inserted into the right child
		err = insertExpect(tree, "r", "2024-01-01T12:00:00Z", "2024-01-01T13:00:00Z", true)
		Tassert(t, err == nil, err)
		// the root interval should move to the left child
		err = expect(tree, "l", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z")
		Tassert(t, err == nil, err)
		// the root interval should be replaced with a new interval
		// that spans the two children
		err = expect(tree, "", "2024-01-01T10:00:00Z", "2024-01-01T13:00:00Z")
		Tassert(t, err == nil, err)

		// insert a interval earlier than all other intervals -- this should
		// cause the new interval to insert at tree.left.left
		err = insertExpect(tree, "ll", "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", true)
		Tassert(t, err == nil, err)
		// tree.left should move to tree.left.right.
		err = expect(tree, "lr", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z")
		Tassert(t, err == nil, err)
		// tree.left should span tree.left.left and tree.left.right
		err = expect(tree, "l", "2024-01-01T09:00:00Z", "2024-01-01T11:00:00Z")
		Tassert(t, err == nil, err)
		// tree should span tree.left and tree.right
		err = expect(tree, "", "2024-01-01T09:00:00Z", "2024-01-01T13:00:00Z")

		// insert an interval between the root and the right child -- this
		// should cause the new interval to insert at tree.right.left
		err = insertExpect(tree, "rl", "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", true)
		Tassert(t, err == nil, err)
		// tree.right should move to tree.right.right
		err = expect(tree, "rr", "2024-01-01T12:00:00Z", "2024-01-01T13:00:00Z")
		Tassert(t, err == nil, err)
		// tree.right should span tree.right.left and tree.right.right
		err = expect(tree, "r", "2024-01-01T11:30:00Z", "2024-01-01T13:00:00Z")
		Tassert(t, err == nil, err)
		// tree should span tree.left and tree.right
		err = expect(tree, "", "2024-01-01T09:00:00Z", "2024-01-01T13:00:00Z")
	*/
}

/*
func TestGenFreeSlots(t *testing.T) {
	tree := NewTree()

	i1000_1100 := insert(tree, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	i1130_1200 := insert(tree, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", true)
	t0900, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	t1300, err := time.Parse(time.RFC3339, "2024-01-01T13:00:00Z")
	Ck(err)

	// genFreeslots returns a channel of free intervals that are
	// discovered by walking the tree in a depth-first manner.  A free
	// interval is a node that has node.Busy() == false.  The minStart
	// and maxEnd times are inclusive.  The duration is exclusive. If
	// first is true, then the intervals are generated in order from
	// the earliest start time to the latest start time.  If first is
	// false, then the intervals are generated in order from the
	// latest start time to the earliest start time.

	slotsChan := tree.genFreeSlots(true)
	var slots []*Interval
	for slot := range slotsChan {
		if slot == nil {
			continue
		}
		slots = append(slots, slot)
	}
	if len(slots) != 3 {
		for _, slot := range slots {
			t.Logf("%v", slot)
		}
		t.Fatalf("Expected 3 slots, got %d", len(slots))
	}

	e0900_1000 := NewInterval(t0900, i1000_1100.Start())
	e1100_1130 := NewInterval(i1000_1100.End(), i1130_1200.Start())
	e1200_1300 := NewInterval(i1130_1200.End(), t1300)

	Tassert(t, slots[0].Equal(e0900_1000), fmt.Sprintf("Expected %s, got %s", e0900_1000, slots[0]))
	Tassert(t, slots[1].Equal(e1100_1130), fmt.Sprintf("Expected %s, got %s", e1100_1130, slots[1]))
	Tassert(t, slots[2].Equal(e1200_1300), fmt.Sprintf("Expected %s, got %s", e1200_1300, slots[2]))

}
*/

/*
func TestConflicts(t *testing.T) {
	tree := NewTree()

	// insert an interval into the tree
	start1, err := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	Ck(err)
	end1, err := time.Parse(time.RFC3339, "2024-01-01T11:00:00Z")
	Ck(err)
	interval1 := NewInterval(start1, end1, true)
	ok := tree.Insert(interval1)
	Tassert(t, ok, "Failed to insert interval")

	// create a new interval that overlaps the first interval
	start2, err := time.Parse(time.RFC3339, "2024-01-01T10:30:00Z")
	Ck(err)
	end2, err := time.Parse(time.RFC3339, "2024-01-01T11:30:00Z")
	Ck(err)
	interval2 := NewInterval(start2, end2, true)

	// Conflicts returns a slice of Intervals
	intervals := tree.Conflicts(interval2)
	Tassert(t, len(intervals) == 1, fmt.Sprintf("Expected 1 conflict, got %d", len(intervals)))
	Tassert(t, intervals[0].Start().Equal(start1), fmt.Sprintf("Expected start1, got %v", intervals[0].Start()))
	Tassert(t, intervals[0].End().Equal(end1), fmt.Sprintf("Expected end1, got %v", intervals[0].End()))
}
*/

/*
func TestMaxGap(t *testing.T) {
	tree := NewTree()

	// insert an interval into the tree
	insert(tree, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z")
	// create a new interval that does not overlap the first interval
	insert(tree, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z")
	Tassert(t, tree.maxGap == 30*time.Minute, fmt.Sprintf("Expected 30 minutes, got %v", tree.maxGap))

	// insert an interval an hour after the second interval
	insert(tree, "2024-01-01T13:00:00Z", "2024-01-01T14:00:00Z")
	Tassert(t, tree.maxGap == 1*time.Hour, fmt.Sprintf("Expected 1 hour, got %v", tree.maxGap))

	// insert an interval in the middle of the free hour
	insert(tree, "2024-01-01T12:10:00Z", "2024-01-01T12:45:00Z")
	// dump(tree, 0)
	Tassert(t, tree.maxGap == 30*time.Minute, fmt.Sprintf("Expected 30 minutes, got %v", tree.maxGap))

	// insert an interval in the middle of the free 30 minutes
	insert(tree, "2024-01-01T11:10:00Z", "2024-01-01T11:20:00Z")
	Tassert(t, tree.maxGap == 15*time.Minute, fmt.Sprintf("Expected 15 minutes, got %v", tree.maxGap))

}

func TestGenSlots(t *testing.T) {
	tree := NewTree()

	i1000_1100 := insert(tree, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	// create a new interval that does not overlap the first interval
	i1130_1200 := insert(tree, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", true)

	t0900, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	t1300, err := time.Parse(time.RFC3339, "2024-01-01T13:00:00Z")
	Ck(err)

	// genslots returns a channel of free intervals that are generated
	// by walking the tree in a depth-first manner.  The minStart and
	// maxEnd times are inclusive.  The duration is exclusive. If first
	// is true, then the intervals are generated in order from the earliest
	// start time to the latest start time.  If first is false, then the
	// intervals are generated in order from the latest start time to the
	// earliest start time.

	slotsChan := tree.genSlots(true, t0900, t1300)
	var slots []*Interval
	for slot := range slotsChan {
		if slot == nil {
			continue
		}
		slots = append(slots, slot)
	}
	if len(slots) != 3 {
		for _, slot := range slots {
			t.Logf("%v", slot)
		}
		t.Fatalf("Expected 3 slots, got %d", len(slots))
	}

	e0900_1000 := NewInterval(t0900, i1000_1100.Start())
	e1100_1130 := NewInterval(i1000_1100.End(), i1130_1200.Start())
	e1200_1300 := NewInterval(i1130_1200.End(), t1300)

	Tassert(t, slots[0].Equal(e0900_1000), fmt.Sprintf("Expected %s, got %s", e0900_1000, slots[0]))
	Tassert(t, slots[1].Equal(e1100_1130), fmt.Sprintf("Expected %s, got %s", e1100_1130, slots[1]))
	Tassert(t, slots[2].Equal(e1200_1300), fmt.Sprintf("Expected %s, got %s", e1200_1300, slots[2]))

}
*/

/*
func TestFreeSlots(t *testing.T) {
	tree := NewTree()

	i1000_1100 := insert(tree, "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	// create a new interval that does not overlap the first interval
	i1130_1200 := insert(tree, "2024-01-01T11:30:00Z", "2024-01-01T12:00:00Z", true)

	t0900, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	t1300, err := time.Parse(time.RFC3339, "2024-01-01T13:00:00Z")
	Ck(err)

	// freeSlots returns at most three intervals:
	//  1. A free interval that starts at the minStart time and ends at the
	//     start of the busy interval in the node.
	//  2. A free interval that starts at the end of the left child's busy
	//     interval and ends at the start of the right child's busy interval.
	//  3. A free interval that starts at the end of the busy interval in the
	//     node and ends at the maxEnd time.
	// insert an interval into the tree

	slots := tree.freeSlots(t0900, t1300)
	Tassert(t, len(slots) == 3, "Expected 3 slots, got %d", len(slots))

	e0900_1000 := NewInterval(t0900, i1000_1100.Start())
	e1100_1130 := NewInterval(i1000_1100.End(), i1130_1200.Start())
	e1200_1300 := NewInterval(i1130_1200.End(), t1300)

	Tassert(t, slots[0].Equal(e0900_1000), fmt.Sprintf("Expected %s, got %s", e0900_1000, slots[0]))
	Tassert(t, slots[1].Equal(e1100_1130), fmt.Sprintf("Expected %s, got %s", e1100_1130, slots[1]))
	Tassert(t, slots[2].Equal(e1200_1300), fmt.Sprintf("Expected %s, got %s", e1200_1300, slots[2]))

}
*/

/*
func TestFree(t *testing.T) {
	tree := NewTree()

	err := insertExpect(tree, "", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	Tassert(t, err == nil, err)
	err = insertExpect(tree, "r", "2024-01-01T12:00:00Z", "2024-01-01T13:00:00Z", true)
	Tassert(t, err == nil, err)
	err = insertExpect(tree, "ll", "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", true)

	// Fetch the first free time between the given start and end times
	// that is the given duration.  The resulting free interval should
	// be 9:30 to 10:00.  As with any interval, the end time is
	// exclusive.
	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	Ck(err)
	// FirstFree walks the tree to find the first free interval that
	// is at least the given duration.  The searchStart and searchEnd
	// times are inclusive.  The duration is exclusive.  The search
	// uses an internal walk() function that recursively walks the
	// tree in a depth-first manner, following the left child first.
	freeInterval := tree.FirstFree(searchStart, searchEnd, 30*time.Minute)
	Tassert(t, freeInterval != nil, "Expected non-nil free interval")
	expectStart, err := time.Parse(time.RFC3339, "2024-01-01T09:30:00Z")
	Ck(err)
	expectEnd, err := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	Ck(err)
	expectInterval := NewInterval(expectStart, expectEnd)
	Tassert(t, freeInterval.Equal(expectInterval), fmt.Sprintf("Expected %s, got %s", expectInterval, freeInterval))

}
*/

func TestFind(t *testing.T) {
	tree := NewTree()

	// insert an interval into the tree -- this should become the left
	// child of the right child of the root node
	err := insertExpect(tree, "rl", "2024-01-01T10:00:00Z", "2024-01-01T11:00:00Z", true)
	Tassert(t, err == nil, err)
	err = insertExpect(tree, "rrrl", "2024-01-01T12:00:00Z", "2024-01-01T13:00:00Z", true)
	Tassert(t, err == nil, err)
	err = insertExpect(tree, "lrl", "2024-01-01T09:00:00Z", "2024-01-01T09:30:00Z", true)
	Tassert(t, err == nil, err)

	searchStart, err := time.Parse(time.RFC3339, "2024-01-01T09:00:00Z")
	Ck(err)
	searchEnd, err := time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	Ck(err)

	// find returns an interval that has the given duration.  The interval
	// starts as early as possible if first is true, and as late as possible
	// if first is false.  The minStart and maxEnd times are inclusive.
	// The duration is exclusive.
	//
	// This function works by walking the tree in a depth-first manner,
	// following the left child first if first is set, otherwise following
	// the right child first.  For each node, it uses freeSlots() to
	// create free intervals.  These intervals are then sorted based on
	// the value of first.  Then they are checked, in order, to see if
	// they have the required duration.  The first one that does
	// is used to create the resulting interval for return.

	// dump(tree, "")
	// find the first free interval that is at least 30 minutes long
	freeInterval := tree.FindFree(true, searchStart, searchEnd, 30*time.Minute)
	Tassert(t, freeInterval != nil, "Expected non-nil free interval")
	expectStart, err := time.Parse(time.RFC3339, "2024-01-01T09:30:00Z")
	Ck(err)
	expectEnd, err := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	Ck(err)
	expectInterval := NewInterval(expectStart, expectEnd, nil)
	Tassert(t, freeInterval.Equal(expectInterval), fmt.Sprintf("Expected %s, got %s", expectInterval, freeInterval))

	// find the last free interval that is at least 30 minutes long
	freeInterval = tree.FindFree(false, searchStart, searchEnd, 30*time.Minute)
	Tassert(t, freeInterval != nil, "Expected non-nil free interval")
	expectStart, err = time.Parse(time.RFC3339, "2024-01-01T17:00:00Z")
	Ck(err)
	expectEnd, err = time.Parse(time.RFC3339, "2024-01-01T17:30:00Z")
	Ck(err)
	expectInterval = NewInterval(expectStart, expectEnd, nil)
	Tassert(t, freeInterval.Equal(expectInterval), fmt.Sprintf("Expected %s, got %s", expectInterval, freeInterval))

}

/*
func TestFindFreeMany(t *testing.T) {
	// This test creates a large tree with a large number of random intervals and then
	// finds free intervals of varying durations.
	rand.Seed(1)
	tree := NewTree()

	// insert several random intervals
	for i := 0; i < 10; i++ {
		start := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
		end := start.Add(time.Duration(rand.Intn(60)) * time.Minute)
		insert(tree, start.Format("2006-01-02T15:04:05"), end.Format("2006-01-02T15:04:05"), true)
	}

	// find a large number of free intervals of varying durations
	for i := 0; i < 100; i++ {
		minStart := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
		maxEnd := minStart.Add(time.Duration(rand.Intn(1440)) * time.Minute)
		duration := time.Duration(rand.Intn(60)) * time.Minute
		first := rand.Intn(2) == 0
		freeInterval := tree.FindFree(first, minStart, maxEnd, duration)
		if freeInterval == nil {
			// sanity check -- try a bunch of times to see if we can find a free interval
			for j := 0; j < 100; j++ {
				start := time.Date(2024, 1, 1, rand.Intn(24), rand.Intn(60), 0, 0, time.UTC)
				end := start.Add(duration)
				ckInterval := NewInterval(start, end, true)
				if tree.Conflicts(ckInterval) == nil {
					t.Logf("Found free interval: %v", ckInterval)
					t.Logf("first: %v, minStart: %v, maxEnd: %v, duration: %v", first, minStart, maxEnd, duration)
					for _, interval := range tree.Intervals() {
						t.Logf("%v", interval)
					}
					t.Fatalf("Expected conflict, got nil")
				}
			}
		}

		if freeInterval.Duration() < duration {
			t.Fatalf("Expected duration of at least %v, got %v", duration, freeInterval.Duration())
		}

		conflicts := tree.Conflicts(freeInterval)
		if conflicts != nil {
			t.Logf("Free interval conflict: %v", freeInterval)
			t.Logf("first: %v, minStart: %v, maxEnd: %v, duration: %v", first, minStart, maxEnd, duration)
			for _, interval := range conflicts {
				t.Logf("%v", interval)
			}
			dump(tree, 0)
			t.Fatalf("Expected free interval, got conflict")
		}

	}
}
*/
