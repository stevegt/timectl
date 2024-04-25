package timectl

import (
	"time"

	. "github.com/stevegt/goadapt"
)

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

	// find free time to fit the new interval -- use the non-locking version
	free := t.findFreePriority(first, minStart, maxEnd, interval.Duration(), interval.Priority())
	if len(free) == 0 {
		return nil, nil, false
	}

	// find and remove any intervals that overlap with the new interval
	start := MaxTime(minStart, free[0].Start())
	end := MinTime(maxEnd, free[len(free)-1].End())
	removed = t.RemoveRange(start, end)

	// adjust the start and end times of the new interval to fit within the found free time
	newStart := start
	newEnd := newStart.Add(interval.Duration())
	newInterval = interval.Clone()
	newInterval.SetStart(newStart)
	newInterval.SetEnd(newEnd)

	// insert the new interval into the tree
	if !t.insert(newInterval) {
		// XXX re-insert removed intervals
		return nil, nil, false
	}

	return newInterval, removed, true
}

func (t *Tree) RemoveRange(start, end time.Time) (removed []Interval) {

	interval := NewInterval(start, end, 0)
	removed = t.conflicts(interval)

	for _, conflict := range removed {
		ok := t.delete(conflict)
		Assert(ok, "failed to delete interval")
	}
	return removed
}
