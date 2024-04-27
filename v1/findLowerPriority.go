package timectl

import (
	"time"
)

// FindLowerPriority searches for the first contiguous set of intervals with lower priority,
// spanning at least a specified duration, starting from either minStart or maxEnd.
func (t *Tree) FindLowerPriority(first bool, minStart, maxEnd time.Time, duration time.Duration, priority float64) []Interval {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := []Interval{}        // To store the final slice of intervals.
	var sumDuration time.Duration // To sum up durations of found intervals.

	// A helper function to accumulate intervals of lower priority.
	var accumulateIntervals func(node *Tree, start time.Time, end time.Time) bool
	accumulateIntervals = func(node *Tree, start time.Time, end time.Time) bool {
		if node == nil || sumDuration >= duration {
			return true // Base case: node is nil or we have enough duration.
		}

		// if the node's minStart is completely after the search range, skip it.
		if node.minStart.After(end) {
			return false
		}

		// if the node's maxEnd is completely before the search range, skip it.
		if node.maxEnd.Before(start) {
			return false
		}

		// if the node's maxPriority is not lower than the required
		// priority, clear the accumulators and return false.
		if node.maxPriority >= priority {
			sumDuration = 0
			result = []Interval{}
			return false
		}

		// Depending on the search direction, recursively accumulate child intervals first.
		if first {
			if accumulateIntervals(node.left, start, end) {
				return true // Stop if already found enough duration.
			}
		} else {
			if accumulateIntervals(node.right, start, end) {
				return true // Stop if already found enough duration.
			}
		}

		// Check this interval if it's within our search range and of lower priority.
		if node.interval.Start().Before(end) && node.interval.End().After(start) && node.interval.Priority() < priority {
			intervalDuration := node.interval.Duration()
			sumDuration += intervalDuration
			result = append(result, node.interval)
		}

		// Continue accumulating intervals based on search direction.
		if first {
			return accumulateIntervals(node.right, start, end)
		} else {
			return accumulateIntervals(node.left, start, end)
		}
	}

	// Kick off accumulation process from the root.
	accumulateIntervals(t, minStart, maxEnd)

	if sumDuration < duration { // Check if we didn't find enough duration.
		return []Interval{} // Return an empty slice in case of failure.
	}

	// Reverse the slice if we were searching from the end to keep intervals in chronological order.
	if !first {
		for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
			result[i], result[j] = result[j], result[i]
		}
	}

	// Return the result slice up to the required duration or all if sumDuration was met or exceeded.
	return result
}
