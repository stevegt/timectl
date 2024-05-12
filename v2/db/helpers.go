package db

import (
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/v2/interval"
)

// Tadd is a test helper function that add an interval to the db
// and returns the interval that was added.  It panics on error.
func Tadd(tx Tx, id uint64, startStr, endStr string, priority float64) *interval.Interval {
	start, err := time.Parse("2006-01-02T15:04:05", startStr)
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", endStr)
	Ck(err)
	iv := interval.NewInterval(id, start, end, priority)
	err = tx.Add(iv)
	Ck(err)
	return iv
}
