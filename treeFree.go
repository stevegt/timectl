package interval

import "time"

// FirstFree finds the first available time slot within the given start and end times
// that is at least of the given duration. It returns nil if no slot is found.
func (t *Tree) FirstFree(startTime, endTime time.Time, duration time.Duration) *Interval {

	var walk func(*Tree, time.Time) *time.Time
	walk = func(node *Tree, currentStart time.Time) *time.Time {
		if node == nil {
			return nil
		}

		// Base case: if current interval starts after currentStart and there's enough room
		if node.interval != nil && node.interval.Start().After(currentStart) && node.interval.Start().Sub(currentStart) >= duration {
			return &currentStart
		}

		leftResult := walk(node.left, currentStart)
		if leftResult != nil {
			return leftResult
		}

		// After visiting left, check if the gap between this node's interval and the next allows for duration
		if node.interval != nil {
			newCurrentStart := node.interval.End()
			if newCurrentStart.Before(endTime) {
				return walk(node.right, newCurrentStart)
			}
		}

		return nil
	}

	freeStartTime := walk(t, startTime)
	if freeStartTime != nil && freeStartTime.Add(duration).Before(endTime) {
		return NewInterval(*freeStartTime, freeStartTime.Add(duration))
	}

	return nil
}
