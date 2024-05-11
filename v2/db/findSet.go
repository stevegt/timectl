package db

import (
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// FindSet returns a contiguous set of intervals that intersect
// with the given start and end time, are lower than the given
// priority, and whose total duration is greater than or equal to the
// given duration.  The first parameter indicates whether the set
// should be the first or last match found within the given time
// range. The results include synthetic free intervals that represent
// the time slots between the intervals.
func (tx *MemTx) FindSet(first bool, minStart, maxEnd time.Time, minDuration time.Duration, maxPriority float64) (set []*interval.Interval, err error) {
	defer Return(&err)

	var candidates []*interval.Interval
	if first {
		candidates, err = tx.FindFwd(minStart, maxEnd, maxPriority)
		Ck(err)
	} else {
		candidates, err = tx.FindBack(minStart, maxEnd, maxPriority)
		Ck(err)
	}

	// spew.Dump(candidates)

	var foundDuration time.Duration
	for _, iv := range candidates {
		if foundDuration >= minDuration {
			return
		}
		if len(set) != 0 {
			// if the previous interval is not contiguous, reset
			prevIv := set[len(set)-1]
			if first && iv.Start.After(prevIv.End) {
				set = nil
				foundDuration = 0
			}
			if !first && iv.End.Before(prevIv.Start) {
				set = nil
				foundDuration = 0
			}
		}
		if len(set) == 0 {
			set = append(set, iv)
			foundDuration = iv.Duration()
			continue
		}
		// we have a contiguous interval; add it to the set
		set = append(set, iv)
		foundDuration += iv.Duration()
	}

	// if we get here, we didn't find a set that meets the criteria
	return nil, nil
}
