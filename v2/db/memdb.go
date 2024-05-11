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

// FindRev returns all intervals that intersect with the given
// given start and end time and are at or lower than the given
// priority.  The results are sorted in descending order by start time.
// The results include synthetic free intervals that represent the
// time slots between the intervals.
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
