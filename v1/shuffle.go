package timectl

import "time"

// Shuffle inserts a new interval into the tree. It finds one or
// more lower-priority intervals using FindFreePriority, removes
// and returns them, adjusts the start and end times of the new
// interval to fit within the found free time, and inserts the new
// interval into the tree. Shuffle returns the newly created interval
// and the removed intervals if successful, otherwise it returns nil.
// The 'first' parameter determines whether to start
// searching for free time at the beginning or end of the given
// time range. Shuffle does not return intervals that are
// marked as free (priority 0) -- it instead adjusts free intervals
// to fill gaps in the tree.
func (t *Tree) Shuffle(first bool, minStart, maxEnd time.Time, interval Interval) (newInterval Interval, removed []Interval, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// find free time to fit the new interval
	free := t.FindFreePriority(first, minStart, maxEnd, interval.Duration(), interval.Priority())
	if len(free) == 0 {
		return nil, nil, false
	}

	// find and remove any lower-priority intervals that overlap with the new interval
	var conflicts []Interval
	for _, f := range free {
		conflicts = append(conflicts, t.Conflicts(f)...)
	}
	for _, conflict := range conflicts {
		if conflict.Priority() < interval.Priority() {
			t.Delete(conflict)
			removed = append(removed, conflict)
		}
	}

	// adjust the start and end times of the new interval to fit within the found free time
	newStart := free[0].Start()
	var durationAccumulated time.Duration = 0
	for _, f := range free {
		durationAccumulated += f.Duration()
		if durationAccumulated >= interval.Duration() {
			break
		}
	}
	newEnd := free[0].Start().Add(interval.Duration())

	newInterval = NewInterval(newStart, newEnd, interval.Priority())

	// insert the new interval into the tree
	if !t.Insert(newInterval) {
		return nil, nil, false
	}

	return newInterval, removed, true
}
