package mem

import (
	"github.com/stevegt/timectl/storage"
	"time"

	"github.com/hashicorp/go-memdb"
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// MemTx is a transaction for the in-memory database.
type MemTx struct {
	tx *memdb.Txn
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
func (tx *MemTx) FindFwdIter(minStart, maxEnd time.Time, maxPriority float64) (iter storage.Iterator, err error) {
	return NewFindIterator(tx, true, minStart, maxEnd, maxPriority)
}

// FindRevIter is the same as FindFwdIter, but it returns the results
// in descending order by start time.
func (tx *MemTx) FindRevIter(minStart, maxEnd time.Time, maxPriority float64) (ivs storage.Iterator, err error) {
	return NewFindIterator(tx, false, minStart, maxEnd, maxPriority)
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
