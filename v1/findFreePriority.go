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
	var accumulatedDuration time.Duration

	// Verify if an interval (or accumulated intervals) can be added to results
	considerAddInterval := func(curEnd time.Time, curPriority float64) bool {
		if tempStart.IsZero() || accumulatedDuration < duration {
			return false
		}
		results = append(results, NewInterval(tempStart, curEnd, curPriority))
		accumulatedDuration = 0
		tempStart = time.Time{}
		return true
	}

	var search func(node *Tree, depth int)
	search = func(node *Tree, depth int) {
		if node == nil {
			return
		}

		if first {
			search(node.left, depth+1)
		} else {
			search(node.right, depth+1)
		}

		if node.leafInterval != nil && node.leafInterval.Start().Before(maxEnd) && node.leafInterval.End().After(minStart) {
			curStart := MaxTime(node.leafInterval.Start(), minStart)
			curEnd := MinTime(node.leafInterval.End(), maxEnd)
			curDuration := curEnd.Sub(curStart)

			if node.leafInterval.Priority() < priority {
				if tempStart.IsZero() {
					tempStart = curStart
				}
				accumulatedDuration += curDuration
				if accumulatedDuration >= duration {
					if considerAddInterval(curEnd, node.leafInterval.Priority()) {
						return // If duration met and added, stop searching further
					}
				}
			} else {
				// Current interval has a priority that is not lower than the search priority
				// Check if accumulated intervals till now meet required duration and add them
				considerAddInterval(curStart, priority)
			}
		}

		if first {
			search(node.right, depth+1)
		} else {
			search(node.left, depth+1)
		}
	}

	search(t, 0)

	// After traversal, check if the last accumulation meets the criteria
	// This handles the case where search ends with accumulated intervals satisfying the duration requirement
	if !tempStart.IsZero() && accumulatedDuration >= duration {
		considerAddInterval(maxEnd, priority) // Use maxEnd, as it's the end of the search range
	}
	
	return results
}
