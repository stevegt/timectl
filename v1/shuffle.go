package timectl

import "time"

// Shuffle inserts a new interval into the tree.  It finds one or
// more lower-priority intervals using findFreePriority, removes
// and returns them, adjusts the start and end times of the new
// interval to fit within the free time, and inserts the new
// interval into the tree.  Shuffle returns the new interval and
// the removed intervals if successful, otherwise it returns nil,
// nil. The 'first' parameter determines whether to start
// searching for free time at the beginning or end of the given
// time range.  Shuffle does not return intervals that are
// marked as free (priority 0) -- it instead adjusts free intervals
// to fill gaps in the tree.
func (t *Tree) Shuffle(first bool, minStart, maxEnd time.Time, interval Interval) (removed []Interval, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// find free time to fit the new interval
	free := t.FindFreePriority(first, minStart, maxEnd, interval.Duration(), interval.Priority())
	if free == nil {
		return nil, false
	}

	// remove any lower-priority intervals that overlap with the free time
	removed := t.RemoveConflicts(free)

	// adjust the start and end times of the new interval to fit within the free time
	newInterval := NewInterval(free.Start(), free.Start().Add(interval.Duration()), interval.Priority())

	// insert the new interval into the tree
	if !t.Insert(newInterval) {
		return nil, false
	}

	return newInterval, removed
}
