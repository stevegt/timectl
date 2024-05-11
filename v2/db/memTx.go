package db

import (
	"time"

	"github.com/hashicorp/go-memdb"
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
	"github.com/stevegt/timectl/util"
)

// MemTx is a transaction for the in-memory database.
type MemTx struct {
	tx *memdb.Txn
}

// NewTx returns a transaction for the database.  If the write
// parameter is true, the transaction is a write transaction.
func (m *Mem) NewTx(write bool) *MemTx {
	return &MemTx{tx: m.memdb.Txn(write)}
}

// Add adds an interval to the database.
func (tx *MemTx) Add(iv *interval.Interval) error {
	// XXX ensure that the interval does not conflict with any existing intervals
	return tx.tx.Insert("interval", iv)
}

// FindFwdIter returns an iterator for the intervals that intersect
// with the given start and end time and are at or lower than the
// given priority.  The results are sorted in ascending order by end
// time.  The results include synthetic free intervals that represent
// the time slots between the intervals.
func (tx *MemTx) FindFwdIter(minStart, maxEnd time.Time, maxPriority float64) (iter Iterator, err error) {
	return NewFindIterator(tx, true, minStart, maxEnd, maxPriority)
}

// FindRevIter is the same as FindFwdIter, but it returns the results
// in descending order by start time.
func (tx *MemTx) FindRevIter(minStart, maxEnd time.Time, maxPriority float64) (ivs Iterator, err error) {
	return NewFindIterator(tx, false, minStart, maxEnd, maxPriority)
}

// FindIterator is an iterator for the Find* functions.
type FindIterator struct {
	boundsIter  memdb.ResultIterator
	fwd         bool
	minStart    time.Time
	maxEnd      time.Time
	maxPriority float64
	visited     *interval.Interval
	queue       []*interval.Interval
}

// NewFindIterator creates a new FindIterator.
func NewFindIterator(tx *MemTx, fwd bool, minStart, maxEnd time.Time, maxPriority float64) (iter *FindIterator, err error) {
	defer Return(&err)

	var boundsIter memdb.ResultIterator
	if fwd {
		boundsIter, err = tx.tx.LowerBound("interval", "end", minStart)
		Ck(err)
	} else {
		boundsIter, err = tx.tx.ReverseLowerBound("interval", "start", maxEnd)
		Ck(err)
	}

	iter = &FindIterator{
		boundsIter:  boundsIter,
		fwd:         fwd,
		minStart:    minStart,
		maxEnd:      maxEnd,
		maxPriority: maxPriority,
	}

	return
}

// Next returns the next interval.
func (iter *FindIterator) Next() *interval.Interval {
	// How this works:  We keep a short queue of intervals we want to
	// return.  If the queue is not empty, we return the first interval
	// in the queue.  Otherwise, we fetch and filter intervals from
	// the underlying bounds iterator, creating free intervals as
	// needed, putting results on the queue.  The queue is the only
	// place we return intervals from.

	// retry until we have something to return
	for {
		// return any queued intervals -- this is the only place we
		// return from, including free intervals and nil
		if len(iter.queue) > 0 {
			iv := iter.queue[0]
			iter.queue = iter.queue[1:]
			return iv
		}

		// get the next interval from the bounds iterator
		obj := iter.boundsIter.Next()
		if obj == nil {
			iter.queue = append(iter.queue, nil)
			continue
		}
		iv := obj.(*interval.Interval)

		// create a free interval between the last-visited interval and the current interval
		var freeStart, freeEnd time.Time
		if iter.fwd {
			// we're iterating forward
			if iter.visited != nil && iv.Start.After(iter.visited.End) {
				// current interval starts after the last-visited interval ends
				// free start time is the previous interval's end time
				freeStart = iter.visited.End
				// free end time is the current interval's start time or the max end time, whichever is earlier
				freeEnd = util.MinTime(iv.Start, iter.maxEnd)
			}
		} else {
			// we're iterating backward
			if iter.visited != nil && iv.End.Before(iter.visited.Start) {
				// current interval ends before the last-visited interval starts
				// free start time is the current interval's end time or the min start time, whichever is later
				freeStart = util.MaxTime(iv.End, iter.minStart)
				// free end time is the last-visited interval's start time
				freeEnd = iter.visited.Start
			}
		}
		// update the last-visited interval
		iter.visited = iv
		// create the free interval
		free := &interval.Interval{
			Start:    freeStart,
			End:      freeEnd,
			Priority: 0,
		}
		// ensure the free interval has a positive duration -- it
		// could be zero if the previous interval ends at the same
		// time as iv.Start
		if free.End.After(free.Start) {
			// we have a valid free interval -- put it in the queue
			iter.queue = append(iter.queue, free)
		}

		if iter.fwd && iv.IsAfterTime(iter.maxEnd) {
			// we're iterating forward and the interval starts on or after the max end time: we are done
			iter.queue = append(iter.queue, nil)
			continue
		} else if !iter.fwd && iv.IsBeforeTime(iter.minStart) {
			// we're iterating backward and the interval ends on or before the min start time: we are done
			iter.queue = append(iter.queue, nil)
			continue
		}

		// If the interval is not within the min start and max end
		// times, skip it. This can happen on the first call to Next()
		// because the LowerBound call returns the first interval
		// that ends on or after the min start time, and the
		// ReverseLowerBound call returns the first interval that
		// starts on or after the max end time.
		if iv.IsBeforeTime(iter.minStart) || iv.IsAfterTime(iter.maxEnd) {
			continue
		}

		// if the interval has a higher priority than the max priority, skip it
		if iv.Priority > iter.maxPriority {
			continue
		}

		// we have a valid interval -- put it in the queue
		iter.queue = append(iter.queue, iv)
	}
}

// FindFwd is a convenience method that returns the results of
// FindFwdIter as a slice.
func (tx *MemTx) FindFwd(minStart, maxEnd time.Time, maxPriority float64) (ivs []*interval.Interval, err error) {
	iter, err := tx.FindFwdIter(minStart, maxEnd, maxPriority)
	Ck(err)
	for {
		iv := iter.Next()
		if iv == nil {
			break
		}
		ivs = append(ivs, iv)
	}
	return
}

// FindRev is a convenience method that returns the results of
// FindRevIter as a slice.
func (tx *MemTx) FindRev(minStart, maxEnd time.Time, maxPriority float64) (ivs []*interval.Interval, err error) {
	iter, err := tx.FindRevIter(minStart, maxEnd, maxPriority)
	Ck(err)
	for {
		iv := iter.Next()
		if iv == nil {
			break
		}
		ivs = append(ivs, iv)
	}
	return
}

// Delete removes an interval from the database.  If the interval
// interval does not exist, it returns an error.
func (tx *MemTx) Delete(iv *interval.Interval) error {
	return tx.tx.Delete("interval", iv)
}

// Commit commits the transaction.
func (tx *MemTx) Commit() {
	tx.tx.Commit()
}

// Abort aborts the transaction.
func (tx *MemTx) Abort() {
	tx.tx.Abort()
}

// Close closes the database.  In the case of an in-memory database,
// this just releases the resources.
func (m *Mem) Close() {
	*m = Mem{}
}
