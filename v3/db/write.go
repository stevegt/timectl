package db

import (
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/v3/interval"
)

// AddStr adds an interval to the database given strings for the start
// and end times.  The strings must be in RFC3339 format.
func AddStr(tx Tx, id uint64, startStr, endStr string, priority float64) (err error) {
	defer Return(&err)

	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)

	iv := interval.NewInterval(id, start, end, priority)
	return tx.Add(iv)
}
