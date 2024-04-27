package timectl

import (
	"time"
)

// FindFreePriority searches for a sequence of contiguous intervals within the specified range
// that have a priority lower than the given priority, and collectively meet
// the required duration. It returns these intervals in chronological order.
func (t *Tree) FindFreePriority(first bool, minStart, maxEnd time.Time, duration time.Duration, targetPriority float64) []Interval {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.findFreePriority(first, minStart, maxEnd, duration, targetPriority)
}

// findFreePriority is a non-locking version of FindFreePriority for
// internal use by locking methods.
func (t *Tree) findFreePriority(first bool, minStart, maxEnd time.Time, duration time.Duration, targetPriority float64) []Interval {

	// XXX

	return nil
}
