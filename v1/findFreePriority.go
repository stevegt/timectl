package timectl

import (
	"time"
)

// FindFreePriority searches the tree to find one or more contiguous intervals
// that are either free or have a priority less than the given priority, and
// which together satisfy the given duration.
func (t *Tree) FindFreePriority(first bool, minStart, maxEnd time.Time, duration time.Duration, priority float64) (intervals []Interval) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Recursive function to traverse and collect intervals.
	var collect func(node *Tree, start time.Time, end time.Time, remaining time.Duration) time.Duration
	collect = func(node *Tree, start time.Time, end time.Time, remaining time.Duration) time.Duration {
		if node == nil || remaining <= 0 {
			return remaining
		}

		// Verify node's interval falls within the time range.
		if node.Interval() != nil && node.Start().Before(maxEnd) && node.End().After(minStart) {
			adjStart := MaxTime(start, node.Start())
			adjEnd := MinTime(end, node.End())

			if node.leafInterval != nil && node.leafInterval.Priority() < priority {
				availDuration := adjEnd.Sub(adjStart)
				if availDuration > remaining {
					availDuration = remaining
				}

				intervals = append(intervals, NewInterval(adjStart, adjStart.Add(availDuration), node.leafInterval.Priority()))
				remaining -= availDuration
				if remaining <= 0 {
					return 0
				}
			}
		}

		// Continue search in children nodes.
		remaining = collect(node.left, start, end, remaining)
		remaining = collect(node.right, start, end, remaining)
		return remaining
	}

	// Start with full search time range and requested duration.
	remaining := collect(t, minStart, maxEnd, duration)

	if remaining == duration {
		return nil // No suitable intervals found.
	}

	return intervals
}
