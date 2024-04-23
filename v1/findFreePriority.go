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

	// XXX implement pseudocode found in test case

	return results
}
