package db

import (
	"time"

	"github.com/stevegt/timectl/interval"
)

// FindSet returns a contiguous set of intervals that intersect
// with the given start and end time, are lower than the given
// priority, and whose total duration is greater than or equal to the
// given duration.  The first parameter indicates whether the set
// should be the first or last match found within the given time
// range.
func (tx *MemTx) FindSet(first bool, minStart, maxEnd time.Time, minDuration time.Duration, maxPriority float64) (ivs []*interval.Interval, err error) {

	return nil, nil
}
