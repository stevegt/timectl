package interval

import "time"

// FirstFree searches for the first available time slot of a given duration within the specified start 
// and end time. It returns the earliest interval of the specified duration that 
// is free (i.e., does not overlap with any existing intervals in the tree).
func (t *Tree) FirstFree(searchStart, searchEnd time.Time, duration time.Duration) *Interval {
    // Early return if the tree is empty, search times are inverted, or duration is non-positive.
    if t == nil || !searchStart.Before(searchEnd) || duration <= 0 {
        return nil
    }

    return t.findFirstFreeRecursive(searchStart, searchEnd, duration)
}

// findFirstFreeRecursive is a helper method that recursively searches for the first free interval.
func (t *Tree) findFirstFreeRecursive(searchStart, searchEnd time.Time, duration time.Duration) *Interval {
    // Base case: If there is no conflict with an existing interval and the duration fits within the search range,
    // then this time slot is free.
    freeInterval := NewInterval(searchStart, searchStart.Add(duration))
    if searchStart.Add(duration).After(searchEnd) {
        return nil // No sufficient slot found within the search constraints
    }

    if len(t.Conflicts(freeInterval)) == 0 { // Check if the interval is indeed free
        return freeInterval
    }

    // If a conflict is found, find the next potential start time by checking the end time of overlapping intervals
    conflictingIntervals := t.Conflicts(NewInterval(searchStart, searchEnd))
    for _, interval := range conflictingIntervals {
        if interval.End().After(searchStart) {
            // Adjust search start to the end of the current conflicting interval
            nextSearchStart := interval.End()
            // Recursively search for free interval with the new start time 
            nextFreeInterval := t.findFirstFreeRecursive(nextSearchStart, searchEnd, duration)
            if nextFreeInterval != nil {
                return nextFreeInterval
            }
        }
    }
    
    return nil
}
