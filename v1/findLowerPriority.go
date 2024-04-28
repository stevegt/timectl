package timectl

import (
	"time"

	. "github.com/stevegt/goadapt"
)

// FindLowerPriority returns a contiguous set of nodes that have a
// lower priority than the given priority.  The start time of the
// first node is on or before minStart, and the end time of the last
// node is on or after maxEnd.  The nodes must total at least the
// given duration, and may be longer.  If first is true, then the
// search starts at minStart and proceeds in order, otherwise the
// search starts at maxEnd and proceeds in reverse order.
func (t *Tree) FindLowerPriority(first bool, searchStart, searchEnd time.Time, duration time.Duration, priority float64) []Interval {
	// get the intervals that overlap the range
	acc := t.accumulate(searchStart, searchEnd)
	// filter the intervals to only include those with a priority less
	// than priority
	low := filter(acc, func(interval Interval) bool {
		return interval.Priority() < priority
	})
	var ordered <-chan Interval
	if first {
		ordered = low
	} else {
		ordered = reverse(low)
	}
	// filter the intervals to only include those that are contiguous
	// for at least duration
	cont := contiguous(ordered, duration)
	if !first {
		cont = reverse(cont)
	}
	res := chan2slice(cont)
	return res
}

func (t *Tree) XXXFindLowerPriority(first bool, minStart, maxEnd time.Time, duration time.Duration, priority float64) []Interval {
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

		Pf("accumulateIntervals: start=%v, end=%v, node.interval=%v, node.minStart=%v, node.maxEnd=%v, node.maxPriority=%v\n",
			start, end, node.interval, node.minStart, node.maxEnd, node.maxPriority)

		// if the node's minStart is completely after the search range, skip it.
		if node.minStart.After(end) {
			return false
		}

		// if the node's maxEnd is completely before the search range, skip it.
		if node.maxEnd.Before(start) {
			return false
		}

		// if the node's maxPriority is not lower than the required
		// priority, clear the accumulators and return false
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
		ckInterval := NewInterval(start, end, 0)
		if ckInterval.Wraps(node.interval) && node.interval.Priority() < priority {
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

// accumulate returns a channel of intervals in the tree that overlap the
// given range of start and end times. The intervals are returned in order
// of start time.
func (t *Tree) accumulate(start, end time.Time) (out <-chan Interval) {

	// filter function to check if an interval overlaps the given range.
	filterFn := func(i Interval) bool {
		return i.OverlapsRange(start, end)
	}

	// XXX replace this with allIntervalsChan(fwd) and add a fwd parameter
	// to accumulate() so we can get the intervals in reverse order.
	allIntervals := t.allIntervals()
	c1 := slice2chan(allIntervals)
	out = filter(c1, filterFn)
	return
}

// slice2chan converts a slice of intervals to a channel of intervals.
func slice2chan(intervals []Interval) <-chan Interval {
	ch := make(chan Interval)
	go func() {
		for _, i := range intervals {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

// chan2slice converts a channel of intervals to a slice of intervals.
func chan2slice(ch <-chan Interval) []Interval {
	intervals := []Interval{}
	for i := range ch {
		intervals = append(intervals, i)
	}
	return intervals
}

// filter filters intervals from a channel of intervals based on a
// filter function and returns a channel of intervals.
func filter(ch <-chan Interval, filterFn func(Interval) bool) <-chan Interval {
	out := make(chan Interval)
	go func() {
		for i := range ch {
			if filterFn(i) {
				out <- i
			}
		}
		close(out)
	}()
	return out
}

// contiguous returns a channel of intervals from the input
// channel that are contiguous and have a total duration of at
// least the given duration.  The intervals may be provided in
// either forward or reverse order, and will be returned in the
// order they are provided.
func contiguous(ch <-chan Interval, duration time.Duration) <-chan Interval {
	out := make(chan Interval)
	go func() {
		var sum time.Duration
		var intervals []Interval
		for i := range ch {
			if len(intervals) == 0 {
				intervals = append(intervals, i)
				sum = i.Duration()
				continue
			}
			okFwd := i.Start().Equal(intervals[len(intervals)-1].End())
			okRev := i.End().Equal(intervals[len(intervals)-1].Start())
			if okFwd || okRev {
				intervals = append(intervals, i)
				sum += i.Duration()
				if sum >= duration {
					for _, i := range intervals {
						out <- i
					}
					close(out)
					return
				}
			} else {
				intervals = []Interval{i}
				sum = i.Duration()
			}
		}
		close(out)
	}()
	return out
}

// reverse reverses the order of intervals in a channel of intervals.
func reverse(ch <-chan Interval) <-chan Interval {
	out := make(chan Interval)
	go func() {
		intervals := []Interval{}
		for i := range ch {
			intervals = append(intervals, i)
		}
		for i := len(intervals) - 1; i >= 0; i-- {
			out <- intervals[i]
		}
		close(out)
	}()
	return out
}
