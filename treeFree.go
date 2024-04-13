package interval

/*

// FirstFree finds the first available time slot within the given start and end times
// that is at least of the given duration. It returns nil if no slot is found.
func (t *Tree) FirstFree(startTime, endTime time.Time, duration time.Duration) *Interval {

	findFreeStart := func(node *Tree, tryStart time.Time) *time.Time {
		// If the current node or its interval is nil, then there are no
		// intervals in this subtree, so we can return the current start time
		if node == nil || node.interval == nil {
			return &tryStart
		}

		busy := node.interval
		busyStart := busy.Start()
		busyEnd := busy.End()



		// Base case: if busyStart - tryStart >= duration, then we have found a free interval
		if busyStart.Sub(tryStart) >= duration {
			return &tryStart
		}

		// try the left subtree
		leftResult := findFreeStart(node.left, tryStart)
		if leftResult != nil {
			return leftResult
		}

		// see if there's a gap between the left and right subtrees
		if node.left != nil && node.right != nil {
			leftEnd := node.left.interval.End()
			rightStart := node.right.interval.Start()
			if rightStart.Sub(leftEnd) >= duration {
				return &leftEnd
			}
		}

		// if busyEnd + duration is after the end time, then we can't find a free interval
		if busyEnd.Add(duration).After(endTime) {
			return nil
		}

		// drill down the right subtree
		return findFreeStart(node.right, busyEnd)


		newCurrentStart := node.interval.End()
		if newCurrentStart.Before(endTime) {
			return findFreeStart(node.right, newCurrentStart)
		}

		return nil
	}

	freeStartTime := findFreeStart(t, startTime)
	if freeStartTime != nil && freeStartTime.Add(duration).Before(endTime) {
		return NewInterval(*freeStartTime, freeStartTime.Add(duration))
	}

	return nil
}
*/
