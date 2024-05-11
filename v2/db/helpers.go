package db

import (
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// Insert is a test helper function that inserts an interval into the db
// and returns the interval that was inserted.  It panics on error.
func Insert(tx Tx, id uint64, startStr, endStr string, priority float64) *interval.Interval {
	start, err := time.Parse("2006-01-02T15:04:05", startStr)
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", endStr)
	Ck(err)
	iv := interval.NewInterval(id, start, end, priority)
	err = tx.Add(iv)
	Ck(err)
	return iv
}
