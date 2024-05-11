package db

import (
	"time"

	"github.com/hashicorp/go-memdb"
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
	"github.com/stevegt/timectl/util"
)

// Mem is an in-memory database.
type Mem struct {
	memdb *memdb.MemDB
}

// NewMemDb creates a new in-memory database.
func NewMem() (mem *Mem, err error) {
	defer Return(&err)

	// Create the DB schema
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"interval": &memdb.TableSchema{
				Name: "interval",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "Id"},
					},
					"start": &memdb.IndexSchema{
						Name:    "start",
						Unique:  true,
						Indexer: &TimeFieldIndex{Field: "Start"},
					},
					"end": &memdb.IndexSchema{
						Name:    "end",
						Unique:  true,
						Indexer: &TimeFieldIndex{Field: "End"},
					},
					"priority": &memdb.IndexSchema{
						Name:    "priority",
						Unique:  false,
						Indexer: &FloatFieldIndex{Field: "Priority"},
					},
				},
			},
		},
	}

	// Create a new data base
	hdb, err := memdb.NewMemDB(schema)
	Ck(err)
	mem = &Mem{memdb: hdb}
	return
}

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

// FindIterator is an iterator for the Find* functions.
type FindIterator struct {
	filterIter memdb.ResultIterator
	fwd        bool
	minStart   time.Time
	maxEnd     time.Time
	mark       time.Time
	queue      []*interval.Interval
}

// NewFindIterator creates a new FindIterator.
func NewFindIterator(tx *MemTx, fwd bool, minStart, maxEnd time.Time, maxPriority float64) (iter *FindIterator, err error) {
	defer Return(&err)
	// filter function returns true if the interval should be skipped
	filter := func(obj interface{}) bool {
		iv := obj.(*interval.Interval)
		// if the interval has a higher priority than the max priority, skip it
		if iv.Priority > maxPriority {
			return true
		}
		// If the interval ends on or before the min start time, skip it.
		// We need this check because the LowerBound function returns the
		// first interval that ends on or after the min start time.
		if iv.IsBeforeTime(minStart) {
			return true
		}
		return false
	}

	var resIter memdb.ResultIterator
	if fwd {
		resIter, err = tx.tx.LowerBound("interval", "end", minStart)
		Ck(err)
	} else {
		resIter, err = tx.tx.ReverseLowerBound("interval", "start", maxEnd)
		Ck(err)
	}

	filterIter := memdb.NewFilterIterator(resIter, filter)

	var mark time.Time
	if fwd {
		mark = minStart
	} else {
		mark = maxEnd
	}

	iter = &FindIterator{
		filterIter: filterIter,
		fwd:        fwd,
		minStart:   minStart,
		maxEnd:     maxEnd,
		mark:       mark,
	}

	return
}

// Next returns the next interval.
func (iter *FindIterator) Next() *interval.Interval {
	// return any queued intervals
	if len(iter.queue) > 0 {
		iv := iter.queue[0]
		iter.queue = iter.queue[1:]
		return iv
	}

	// get the next interval from the filter iterator
	obj := iter.filterIter.Next()
	if obj == nil {
		return nil
	}
	iv := obj.(*interval.Interval)

	// create a free interval between the last-visited interval and the current interval
	var freeStart, freeEnd time.Time
	if iter.fwd && iv.Start.After(iter.mark) {
		// iv starts after the last-visited interval ends
		// free start time is the previous interval's end time
		freeStart = iter.mark
		// free end time is the current interval's start time or the max end time, whichever is earlier
		// XXX simplify to iter.mark?
		freeEnd = util.MinTime(iv.Start, iter.maxEnd)
		iter.mark = iv.End
	} else if !iter.fwd && iv.End.Before(iter.mark) {
		// iv ends before the last-visited interval starts
		// free start time is the current interval's end time or the min start time, whichever is later
		// XXX simplify to iter.mark?
		freeStart = util.MaxTime(iv.End, iter.minStart)
		// free end time is the previous interval's start time
		freeEnd = iter.mark
		iter.mark = iv.Start
	}
	free := &interval.Interval{
		Start:    freeStart,
		End:      freeEnd,
		Priority: 0,
	}
	// ensure the free interval has a positive duration -- it could be zero if the previous interval ends at the same time as iv.Start
	if free.End.After(free.Start) {
		// add iv to the queue 'cause we're returning free this time
		iter.queue = append(iter.queue, iv)
		return free
	}

	if iter.fwd && iv.IsAfterTime(iter.maxEnd) {
		// the interval starts on or after the max end time: we are done.
		return nil
	} else if !iter.fwd && iv.IsBeforeTime(iter.minStart) {
		// the interval ends on or before the min start time: we are done.
		return nil
	}

	return iv
}

// FindFwd returns all intervals that intersect with the given
// given start and end time and are at or lower than the given
// priority.  The results are sorted in ascending order by end time.
// The results include synthetic free intervals that represent the
// time slots between the intervals.
func (tx *MemTx) FindFwd(minStart, maxEnd time.Time, maxPriority float64) (ivs []*interval.Interval, err error) {
	iter, err := tx.tx.LowerBound("interval", "end", minStart)
	Ck(err)
	prevEnd := minStart
	for {
		obj := iter.Next()
		if obj == nil {
			break
		}
		iv := obj.(*interval.Interval)

		// create a free interval between the previous interval and the current interval
		if iv.Start.After(prevEnd) {
			free := &interval.Interval{
				// free start time is the previous interval's end time
				Start: prevEnd,
				// free end time is the current interval's start time
				// or the max end time, whichever is earlier
				End:      util.MinTime(iv.Start, maxEnd),
				Priority: 0,
			}
			// ensure the free interval has a positive duration --
			// it could be zero if the previous interval ends at the
			// same time as iv.Start
			if free.End.After(free.Start) {
				ivs = append(ivs, free)
			}
		}
		prevEnd = iv.End

		// if the interval has a higher priority than the max priority, skip it
		if iv.Priority > maxPriority {
			continue
		}
		// If the interval ends on or before the min start time, skip it.
		// We need this check because the LowerBound function returns the
		// first interval that ends on or after the min start time.
		if iv.IsBeforeTime(minStart) {
			continue
		}
		// If the interval starts on or after the max end time, we are done.
		if iv.IsAfterTime(maxEnd) {
			break
		}

		ivs = append(ivs, iv)
	}
	return
}

// FindRev is the same as FindFwd, but it returns the results in
// descending order by start time.
func (tx *MemTx) FindRev(minStart, maxEnd time.Time, maxPriority float64) (ivs []*interval.Interval, err error) {
	iter, err := tx.tx.ReverseLowerBound("interval", "start", maxEnd)
	Ck(err)
	prevStart := maxEnd
	for {
		obj := iter.Next()
		if obj == nil {
			break
		}
		iv := obj.(*interval.Interval)

		// create a free interval between the previous interval and the current interval
		// (remember that we are iterating in descending order, so
		// "previous" means later in time)
		if iv.End.Before(prevStart) {
			free := &interval.Interval{
				// free start time is the current interval's end time
				// or the min start time, whichever is later
				Start: util.MaxTime(iv.End, minStart),
				// free end time is the previous interval's start time
				End:      prevStart,
				Priority: 0,
			}
			// ensure the free interval has a positive duration --
			// it could be zero if the previous interval starts at the
			// same time as iv.End
			if free.End.After(free.Start) {
				ivs = append(ivs, free)
			}
		}
		prevStart = iv.Start

		// if the interval has a higher priority than the max priority, skip it
		if iv.Priority > maxPriority {
			continue
		}
		// If the interval ends on or before the min start time, we are done.
		// We need this check because the ReverseLowerBound function returns the
		// first interval that starts on or after the max end time.
		if iv.IsBeforeTime(minStart) {
			break
		}
		// If the interval starts on or after the max end time, skip it.
		if iv.IsAfterTime(maxEnd) {
			continue
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
