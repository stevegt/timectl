package db

import (
	"time"

	"github.com/hashicorp/go-memdb"
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
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

// Find returns all intervals that intersect with the given
// given start and end time and are at or lower than the given
// priority.  The results are sorted in ascending order by end time.
func (tx *MemTx) Find(minStart, maxEnd time.Time, maxPriority float64) (ivs []*interval.Interval, err error) {
	defer Return(&err)

	return tx.FindAscending(minStart, maxEnd, maxPriority)
}

// FindAscending returns all intervals that intersect with the given
// given start and end time and are at or lower than the given
// priority.  The results are sorted in ascending order by end time.
func (tx *MemTx) FindAscending(minStart, maxEnd time.Time, maxPriority float64) (ivs []*interval.Interval, err error) {
	iter, err := tx.tx.LowerBound("interval", "end", minStart)
	Ck(err)
	for {
		obj := iter.Next()
		if obj == nil {
			break
		}
		iv := obj.(*interval.Interval)
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
		if iv.Priority <= maxPriority {
			ivs = append(ivs, iv)
		}
	}
	return
}

// FindDescending returns all intervals that intersect with the given
// given start and end time and are at or lower than the given
// priority.  The results are sorted in descending order by start time.
func (tx *MemTx) FindDescending(minStart, maxEnd time.Time, maxPriority float64) (ivs []*interval.Interval, err error) {
	iter, err := tx.tx.ReverseLowerBound("interval", "start", maxEnd)
	Ck(err)
	for {
		obj := iter.Next()
		if obj == nil {
			break
		}
		iv := obj.(*interval.Interval)
		// If the interval ends on or before the min start time, we are done.
		if iv.IsBeforeTime(minStart) {
			break
		}
		// If the interval starts on or after the max end time, skip it.
		if iv.IsAfterTime(maxEnd) {
			continue
		}
		if iv.Priority <= maxPriority {
			ivs = append(ivs, iv)
		}
	}
	return
}

// Delete removes an interval from the database.  If the interval
// interval does not exist, it returns an error.
func (tx *MemTx) Delete(iv *interval.Interval) error {
	return tx.tx.Delete("interval", iv)
}
