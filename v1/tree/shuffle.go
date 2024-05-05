package tree

import (
	"fmt"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
	"github.com/stevegt/timectl/util"
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
func (t *Node) Shuffle(first bool, minStart, maxEnd time.Time, iv interval.Interval) (newIv interval.Interval, removed []interval.Interval, err error) {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	// ensure the new interval is busy
	if iv.Priority() == 0 {
		return nil, nil, fmt.Errorf("new interval must have a priority greater than 0")
	}

	// find time to fit the new interval
	lower := t.FindLowerPriority(first, minStart, maxEnd, iv.Duration(), iv.Priority())
	if len(lower) == 0 {
		return nil, nil, fmt.Errorf("no free slots found between %s and %s below priority %f", minStart, maxEnd, iv.Priority())
	}

	// remove and hold onto the found intervals
	for _, node := range lower {
		iv := t.free(node)
		removed = append(removed, iv)
	}

	// adjust the start and end times of the new interval to fit within the found free time
	var newStart, newEnd time.Time
	if first {
		newStart = util.MaxTime(minStart, lower[0].Start())
		newEnd = newStart.Add(iv.Duration())
	} else {
		newEnd = util.MinTime(maxEnd, lower[len(lower)-1].End())
		newStart = newEnd.Add(-iv.Duration())
	}
	iv.SetStart(newStart)
	iv.SetEnd(newEnd)

	// insert the new interval into the tree
	if !t.Insert(iv) {
		// XXX re-insert removed intervals or always return a new tree
		// from functions that modify the tree
		Pf("removed: %v\n", removed)
		return nil, nil, fmt.Errorf("failed to insert new interval")
	}

	return newIv, removed, nil
}
