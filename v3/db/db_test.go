package db_test

import (
	"time"

	"github.com/stevegt/goadapt"

	// . "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/v3/db"
	"github.com/stevegt/timectl/v3/db/mem"
)

var (
	Ck = goadapt.Ck
	Pf = goadapt.Pf
	Pl = goadapt.Pl
)

func ExampleFindSet() {
	memdb, err := mem.NewMem()
	Ck(err)
	tx := memdb.NewTx(true)

	// add several intervals
	db.Tadd(tx, 5, "2024-01-01T08:00:00", "2024-01-01T09:00:00", 1.0)
	db.Tadd(tx, 10, "2024-01-01T09:00:00", "2024-01-01T10:00:00", 2.0)
	db.Tadd(tx, 20, "2024-01-01T10:00:00", "2024-01-01T11:00:00", 3.0)
	db.Tadd(tx, 30, "2024-01-01T11:00:00", "2024-01-01T12:00:00", 2.0)
	db.Tadd(tx, 40, "2024-01-01T12:00:00", "2024-01-01T12:45:00", 1.0)
	// note the 15 minute gap here
	db.Tadd(tx, 50, "2024-01-01T13:00:00", "2024-01-01T14:00:00", 1.0)
	db.Tadd(tx, 60, "2024-01-01T14:00:00", "2024-01-01T15:00:00", 1.0)

	// find the first set of intervals that are within a time range,
	// have a priority less than or equal to 1.0, and have a total
	// duration of at least 90 minutes
	start, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T08:00:00")
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T15:00:00")
	Ck(err)
	ivs, err := db.FindSet(tx, true, start, end, 90*time.Minute, 1.0)
	Pf("FindSet() fwd returned %v intervals:\n", len(ivs))
	Pf("(Note how the 15 minute gap between intervals 40 and 50 got filled\n")
	Pf("by a free interval with Id 0 and priority 0.)\n\n")
	for _, iv := range ivs {
		Pf("Id %v Start %v End %v Priority %v\n", iv.Id, iv.Start, iv.End, iv.Priority)
	}
	Pl()

	// find the last set of intervals that are within a time range,
	// have a priority less than or equal to 1.0, and have a total
	// duration of at least 90 minutes
	ivs, err = db.FindSet(tx, false, start, end, 90*time.Minute, 1.0)
	Pf("FindSet() rev returned %v intervals:\n\n", len(ivs))
	for _, iv := range ivs {
		Pf("Id %v Start %v End %v Priority %v\n", iv.Id, iv.Start, iv.End, iv.Priority)
	}

	// Output:
	// FindSet() fwd returned 3 intervals:
	// (Note how the 15 minute gap between intervals 40 and 50 got filled
	// by a free interval with Id 0 and priority 0.)
	//
	// Id 40 Start 2024-01-01 12:00:00 +0000 UTC End 2024-01-01 12:45:00 +0000 UTC Priority 1
	// Id 0 Start 2024-01-01 12:45:00 +0000 UTC End 2024-01-01 13:00:00 +0000 UTC Priority 0
	// Id 50 Start 2024-01-01 13:00:00 +0000 UTC End 2024-01-01 14:00:00 +0000 UTC Priority 1
	//
	// FindSet() rev returned 2 intervals:
	//
	// Id 60 Start 2024-01-01 14:00:00 +0000 UTC End 2024-01-01 15:00:00 +0000 UTC Priority 1
	// Id 50 Start 2024-01-01 13:00:00 +0000 UTC End 2024-01-01 14:00:00 +0000 UTC Priority 1

}
