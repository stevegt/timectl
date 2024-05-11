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
						Indexer: &memdb.UintFieldIndex{Field: "XXXId"},
					},
					"start": &memdb.IndexSchema{
						Name:    "start",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "Start"},
					},
					"end": &memdb.IndexSchema{
						Name:    "end",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "End"},
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
func (tx *MemTx) Add(iv interval.Interval) error {
	return tx.tx.Insert("interval", iv)
}

// Get returns an interval from the database.
func (tx *MemTx) Get(minStart, maxEnd time.Time, maxPriority float64) (ivs []interval.Interval, err error) {
	defer Return(&err)
	iter, err := tx.tx.LowerBound("interval", "start", minStart)
	Ck(err)
	for {
		obj := iter.Next()
		if obj == nil {
			break
		}
		iv := obj.(interval.Interval)
		if iv.End().Before(maxEnd) && iv.Priority() <= maxPriority {
			ivs = append(ivs, iv)
		}
	}
	return
}
