package interval

// Tree represents an interval tree.
// For this simplified implementation, we'll not implement a full interval tree.
// Instead, we will use a slice to demonstrate the concept.
type Tree struct {
	intervals []*Interval // Change from value to pointer to match the Interval pointer used in tests
}

// NewTree creates and returns a new Tree.
func NewTree() *Tree {
	return &Tree{}
}

// Insert adds a new interval to the tree.
// It does not check for duplicates or overlaps; this method simply adds the interval.
func (t *Tree) Insert(interval *Interval) { // Accept a pointer to Interval
	t.intervals = append(t.intervals, interval)
}

// Conflicts checks if the given interval conflicts with any interval in the tree.
// It returns a slice of intervals that conflict.
// This is a very basic implementation for demonstration purposes.
func (t *Tree) Conflicts(interval *Interval) []*Interval { // Return slice of pointer to Interval
	var conflicts []*Interval
	for _, i := range t.intervals {
		if i.Conflicts(interval) {
			conflicts = append(conflicts, i)
		}
	}
	return conflicts
}

