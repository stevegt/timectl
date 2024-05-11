package db

import (
	"github.com/hashicorp/go-memdb"
	. "github.com/stevegt/goadapt"
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
