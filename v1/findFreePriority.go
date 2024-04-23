package timectl

import (
	"time"
)

// FindFreePriority finds a sequence of contiguous intervals within the specified range
// that are free or have a priority lower than the given priority, and collectively meet
// the required duration. It returns these intervals as a slice.
func (t *Tree) FindFreePriority(first bool, minStart, maxEnd time.Time, duration time.Duration, priority float64) []Interval {
	t.mu.RLock()
	defer t.mu.RUnlock()

	results := make([]Interval, 0)
	var tempStart time.Time
	accumulatedDuration := time.Duration(0)

	// Helper to add an interval to results if it satisfies conditions
	addIntervalIfNeeded := func(intervalEnd time.Time) {
		if tempStart.IsZero() {
			return // No interval to add
		}

		// First check accumulated duration against required duration.
		if accumulatedDuration >= duration {
			results = append(results, NewInterval(tempStart, intervalEnd, priority))
		}

		// Reset accumulators
		tempStart = time.Time{}
		accumulatedDuration = 0
	}

	// Recursive function to traverse tree and find suitable intervals
	var search func(node *Tree)
	search = func(node *Tree) {
		if node == nil {
			return
		}

		if first {
			search(node.left)
		} else {
			search(node.right)
		}

		// Process current node
		if node.leafInterval != nil {
			currentInterval := node.leafInterval
			if currentInterval.Priority() <= priority {
				// Check if this interval can contribute to the required duration
				curStart := MaxTime(currentInterval.Start(), minStart)
				curEnd := MinTime(currentInterval.End(), maxEnd)
				if curStart.Before(curEnd) {
					if tempStart.IsZero() {
						tempStart = curStart
					}
					accumulatedDuration += curEnd.Sub(curStart)
					if accumulatedDuration >= duration {
						addIntervalIfNeeded(curEnd)
					}
				}
			} else {
				// Current interval has higher priority. Check if we have collected enough duration.
				addIntervalIfNeeded(currentInterval.Start())
			}
		}

		if first {
			search(node.right)
		} else {
			search(node.left)
		}
	}

	search(t)

	// Check at the end of traversal if there's an unfinished interval accumulation
	if !tempStart.IsZero() && accumulatedDuration < duration {
		// If accumulated duration at the end is insufficient, discard it by excluding the addition
	}

	return results
}
