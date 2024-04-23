package timectl

import (
	"time"
)

// FindFreePriority searches for a sequence of contiguous intervals within the specified range
// that have a priority lower than the given priority, and collectively meet
// the required duration. It returns these intervals in chronological order.
func (t *Tree) FindFreePriority(first bool, minStart, maxEnd time.Time, duration time.Duration, priority float64) []Interval {
	t.mu.RLock()
	defer t.mu.RUnlock()

	candidateIntervals := []Interval{}
	var accumulatedDuration time.Duration

	checkAndAppend := func(interval Interval) {
		if interval.Start().Before(minStart) || interval.End().After(maxEnd) || interval.Priority() >= priority {
			return // Outside of search range or does not satisfy priority constraint
		}

		if len(candidateIntervals) == 0 || interval.Start().Equal(candidateIntervals[len(candidateIntervals)-1].End()) {
			// Interval is continuous with the last one or is the first interval
			candidateIntervals = append(candidateIntervals, interval)
			accumulatedDuration += interval.End().Sub(interval.Start())
			if accumulatedDuration >= duration {
				// Enough intervals found; stop adding more
				return
			}
		} else {
			// Non-continuous interval, reset candidates
			candidateIntervals = []Interval{interval}
			accumulatedDuration = interval.End().Sub(interval.Start())
		}
	}

	var walk func(node *Tree)
	walk = func(node *Tree) {
		if node == nil || accumulatedDuration >= duration {
			return // Base case or enough duration found
		}
		if first {
			// Traverse in chronological order
			walk(node.left)
			if node.leafInterval != nil {
				checkAndAppend(node.leafInterval)
			}
			walk(node.right)
		} else {
			// Traverse in reverse chronological order
			walk(node.right)
			if node.leafInterval != nil {
				checkAndAppend(node.leafInterval)
			}
			walk(node.left)
		}
	}

	walk(t)

	if accumulatedDuration < duration {
		// Not enough continuous time found; discard partial results
		return nil
	}

	// Return the sequence of intervals that meet the criteria
	return candidateIntervals
}
