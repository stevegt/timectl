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

	var accumulateIntervals func(node *Tree, accumulated []Interval, accumulatedDuration time.Duration) ([]Interval, time.Duration)

	accumulateIntervals = func(node *Tree, accumulated []Interval, accumulatedDuration time.Duration) ([]Interval, time.Duration) {
		if node == nil || accumulatedDuration >= duration {
			return accumulated, accumulatedDuration
		}

		if first {
			accumulated, accumulatedDuration = accumulateIntervals(node.left, accumulated, accumulatedDuration)
			if node.leafInterval != nil && (node.leafInterval.Priority() < targetPriority) && (node.leafInterval.Start().After(minStart) || node.leafInterval.Start().Equal(minStart)) && node.leafInterval.End().Before(maxEnd) {
				if len(accumulated) == 0 || accumulated[len(accumulated)-1].End().Equal(node.leafInterval.Start()) {
					accumulated = append(accumulated, node.leafInterval)
					accumulatedDuration += node.leafInterval.End().Sub(node.leafInterval.Start())
				} else {
					accumulated = []Interval{node.leafInterval}
					accumulatedDuration = node.leafInterval.End().Sub(node.leafInterval.Start())
				}
			}
			accumulated, accumulatedDuration = accumulateIntervals(node.right, accumulated, accumulatedDuration)
		} else {
			accumulated, accumulatedDuration = accumulateIntervals(node.right, accumulated, accumulatedDuration)
			if node.leafInterval != nil && (node.leafInterval.Priority() < targetPriority) && (node.leafInterval.Start().After(minStart) || node.leafInterval.Start().Equal(minStart)) && node.leafInterval.End().Before(maxEnd) {
				if len(accumulated) == 0 || accumulated[len(accumulated)-1].Start().Equal(node.leafInterval.End()) {
					prepended := make([]Interval, 1, len(accumulated)+1)
					prepended[0] = node.leafInterval
					accumulated = append(prepended, accumulated...)
					accumulatedDuration += node.leafInterval.End().Sub(node.leafInterval.Start())
				} else {
					accumulated = []Interval{node.leafInterval}
					accumulatedDuration = node.leafInterval.End().Sub(node.leafInterval.Start())
				}
			}
			accumulated, accumulatedDuration = accumulateIntervals(node.left, accumulated, accumulatedDuration)
		}

		return accumulated, accumulatedDuration
	}

	foundIntervals, foundDuration := accumulateIntervals(t, nil, 0)
	if foundDuration >= duration {
		return foundIntervals
	}

	return nil
}
