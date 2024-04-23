package timectl

import (
	"time"
)

// Correcting the access to properties and methods based on the provided error feedback.

// FindFreePriority finds intervals within the specified time range that are free
// or have a priority lower than the given priority value and satisfy
// the specified duration. It considers the priority to determine if an interval
// is "free enough". This method leverages the structure and logic already
// present in the codebase, thus assuming Tree and Interval are correctly defined previously.
func (t *Tree) FindFreePriority(first bool, minStart, maxEnd time.Time, duration time.Duration, priority float64) []Interval {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Setup placeholders for results and a recursive search function.
	var results []Interval
	var search func(node *Tree, required time.Duration) bool

	search = func(node *Tree, required time.Duration) bool {
		if node == nil || required <= 0 {
			return false
		}

		start := maxTime(node.Interval().Start(), minStart)
		end := minTime(node.Interval().End(), maxEnd)
		currentDuration := minDuration(end.Sub(start), required)

		// Check for sufficient priority and room.
		if start.Before(end) && node.Interval().Priority() <= priority && currentDuration >= required {
			results = append(results, NewInterval(start, start.Add(required), node.Interval().Priority()))
			required -= currentDuration
			return required <= 0
		}

		// Recursively search children.
		if first {
			if search(node.left, required) {
				return true
			}
			return search(node.right, required)
		} else {
			if search(node.right, required) {
				return true
			}
			return search(node.left, required)
		}
	}

	search(t, duration)
	return results
}

// Assuming these utility functions are defined elsewhere in your codebase or are provided here for completion.
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if b.Before(a) {
		return b
	}
	return a
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}


