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
	var collect func(node *Tree, start time.Time, end time.Time) bool
	collect = func(node *Tree, start time.Time, end time.Time) bool {
		if node == nil || duration <= 0 {
			return false
		}

		// Adjust search range based on the node's interval.
		start = MaxTime(start, node.Start())
		end = MinTime(end, node.End())
		if end.Sub(start) < duration {
			return false
		}

		if node.leafInterval != nil && node.leafInterval.Priority() < priority {
			// Found a suitable interval. Adjust its duration and add it to the results.
			if !first && duration < end.Sub(start) {
				start = end.Add(-duration)
			} else if duration < end.Sub(start) {
				end = start.Add(duration)
			}
			interval := NewInterval(start, end, node.leafInterval.Priority())
			intervals = append(intervals, interval)
			duration -= interval.Duration()
			return true
		}

		if first {
			if collect(node.left, start, end) {
				return true // Found a suitable interval in the left subtree.
			}
			return collect(node.right, start, end) // Continue searching in the right subtree.
		}

		if collect(node.right, start, end) {
			return true // Found a suitable interval in the right subtree.
		}
		return collect(node.left, start, end) // Continue searching in the left subtree.
	}

	minStart, maxEnd = MaxTime(t.Start(), minStart), MinTime(t.End(), maxEnd)

	if !collect(t, minStart, maxEnd) && len(intervals) == 0 {
		// If no intervals were found that meet the criteria, return nil.
		return nil
	}

	return intervals
}
