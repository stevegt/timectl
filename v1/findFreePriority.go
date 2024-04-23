package timectl

import (
	"time"

	. "github.com/stevegt/goadapt"
)

// FindFreePriority works similarly to FindFree, but it returns a
// contiguous set of intervals that are either free or have a lower
// priority than the given priority.  The intervals are returned in
// order of start time.  The minStart and maxEnd times are inclusive.
func (t *Tree) FindFreePriority(first bool, minStart, maxEnd time.Time, duration time.Duration, priority float64) (intervals []Interval) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	Pf("FindFreePriority: first: %v minStart: %v maxEnd: %v duration: %v priority: %v\n", first, minStart, maxEnd, duration, priority)

	if t.left == nil && t.right == nil {
		// this is a leaf node
		interval := t.interval()
		if interval.Priority() >= priority {
			// this interval has higher priority than we're looking for
			Pf("priority too high for interval: %v\n", interval)
			return nil
		}
		intervals = append(intervals, interval)
		Pf("returning interval: %v\n", interval)
		return
	}

	var children []*Tree
	var start, end time.Time
	if first {
		children = []*Tree{t.left, t.right}
	} else {
		children = []*Tree{t.right, t.left}
	}

	needDuration := duration
	done := false
	for _, child := range children {
		if child == nil {
			continue
		}
		start = MaxTime(minStart, child.Start())
		end = MinTime(child.End(), maxEnd)
		span := end.Sub(start)
		need := needDuration
		if span < need {
			need = span
		}
		candidates := child.FindFreePriority(first, start, end, need, priority)
		for _, candidate := range candidates {
			Pf("candidate: %v\n", candidate)
			if needDuration <= 0 {
				Pf("needDuration <= 0\n")
				done = true
				break
			}
			// find the intersection of the candidate interval and the
			// start/end range
			start = MaxTime(start, candidate.Start())
			end = MinTime(candidate.End(), end)
			duration := end.Sub(start)
			if duration < 0 {
				// candidate interval is outside of minStart/maxEnd range
				if first {
					if candidate.Start().After(maxEnd) {
						done = true
						Pf("candidate.Start().After(maxEnd)\n")
						break
					}
				} else {
					if candidate.End().Before(minStart) {
						done = true
						Pf("candidate.End().Before(minStart)\n")
						break
					}
				}
				// we're not yet at the minStart/maxEnd range, so continue
				Pf("candidate interval not yet at minStart/maxEnd range\n")
				continue
			}
			Pf("adding candidate to intervals: %v\n", candidate)
			needDuration -= duration
			Pf("needDuration: %v\n", needDuration)
			intervals = append(intervals, candidate)
		}
		if done {
			Pf("done\n")
			break
		}
	}
	Pf("returning intervals at end: %v\n", intervals)
	return intervals
}
