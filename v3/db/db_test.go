package db_test

import (
	"time"

	"github.com/stevegt/goadapt"

	// . "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/v3/db"
	"github.com/stevegt/timectl/v3/db/mem"
	"github.com/stevegt/timectl/v3/interval"
)

var (
	Ck = goadapt.Ck
	Pf = goadapt.Pf
	Pl = goadapt.Pl
)

func ExampleFindSet() {
	// create a new in-memory database and transaction
	memdb, err := mem.NewMem()
	// we're using the goadapt.Ck() helper function to check for
	// errors -- you could also use the idiomatic `if err != nil { ... }`
	Ck(err)
	tx := memdb.NewTx(true)

	// add several intervals -- here we're using the Tadd() test
	// helper method for brevity; normally you would use tx.Add()
	// directly to add intervals, checking errors etc.
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
	// we're using the goadapt.Pf() helper function to print output
	// here -- you could also use the idiomatic `fmt.Printf()`
	Pf("FindSet() fwd returned %v intervals:\n", len(ivs))
	Pf("(Note how the 15 minute gap between intervals 40 and 50 got filled\n")
	Pf("by a free interval with Id 0 and priority 0.)\n\n")
	for _, iv := range ivs {
		Pf("Id %v Start %v End %v Priority %v\n", iv.Id, iv.Start, iv.End, iv.Priority)
	}
	// we're using the goadapt.Pl() helper function to print a newline
	// here -- you could also use the idiomatic `fmt.Println()`
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

func ExampleConflicts() {
	// add an in-memory database and transaction
	memdb, err := mem.NewMem()
	Ck(err)
	tx := memdb.NewTx(true)

	// add several intervals -- here we're using the AddStr() wrapper
	// for tx.Add()
	err = db.AddStr(tx, 10, "2024-01-01T08:00:00", "2024-01-01T09:00:00", 1.0)
	Ck(err)
	err = db.AddStr(tx, 20, "2024-01-01T09:00:00", "2024-01-01T10:00:00", 2.0)
	Ck(err)
	// note the gap between interval 20 and 30 -- this will be a free
	// interval that we'll ignore in the conflicts check
	err = db.AddStr(tx, 30, "2024-01-01T11:00:00", "2024-01-01T12:00:00", 2.0)
	Ck(err)

	// check for conflicts with a new interval that overlaps with
	// existing non-free intervals
	badIv, err := interval.NewIntervalStr(40, "2024-01-01T08:30:00", "2024-01-01T09:30:00", 1.0)
	Ck(err)
	conflicts, err := db.Conflicts(tx, badIv)
	Ck(err)
	if conflicts {
		Pf("Conflicts with existing intervals: %v\n\n", badIv)
	}

	// check for conflicts with a new interval that does not overlap
	// with existing non-free intervals
	goodIv, err := interval.NewIntervalStr(50, "2024-01-01T10:00:00", "2024-01-01T11:00:00", 1.0)
	Ck(err)
	conflicts, err = db.Conflicts(tx, goodIv)
	if !conflicts {
		Pf("No conflicts with existing intervals: %v\n", goodIv)
	}

}
