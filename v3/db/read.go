package db

import (
	"math"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/v3/interval"
)

// FindSet returns a contiguous set of intervals that intersect
// with the given start and end time, are lower than the given
// priority, and whose total duration is greater than or equal to the
// given duration.  The first parameter indicates whether the set
// should be the first or last match found within the given time
// range. The results include synthetic free intervals that represent
// the time slots between the intervals.
func FindSet(tx Tx, first bool, minStart, maxEnd time.Time, minDuration time.Duration, maxPriority float64) (set []*interval.Interval, err error) {
	defer Return(&err)

	var candidates Iterator
	if first {
		candidates, err = tx.FindFwdIter(minStart, maxEnd, maxPriority)
		Ck(err)
	} else {
		candidates, err = tx.FindRevIter(minStart, maxEnd, maxPriority)
		Ck(err)
	}

	// spew.Dump(candidates)

	var foundDuration time.Duration
	for {
		if foundDuration >= minDuration {
			// we have a set that meets the criteria
			return
		}

		iv := candidates.Next()
		if iv == nil {
			// we didn't find a set that meets the criteria
			return nil, nil
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
			// start a new set
			set = append(set, iv)
			foundDuration = iv.Duration()
			continue
		}
		// we have a contiguous interval; add it to the set
		set = append(set, iv)
		foundDuration += iv.Duration()
	}

}

// Conflicts returns true if the given interval conflicts with any
// existing intervals in the database.
func Conflicts(tx Tx, iv *interval.Interval) (conflicts bool, err error) {
	defer Return(&err)

	// find all intervals that intersect with the given interval
	iter, err := tx.FindFwdIter(iv.Start, iv.End, math.MaxFloat64)
	Ck(err)

	// any non-zero priority interval that intersects with the given
	// interval is a conflict
	for {
		found := iter.Next()
		if found == nil {
			break
		}
		if found.Priority != 0 {
			return true, nil
		}
	}

	return false, nil
}
