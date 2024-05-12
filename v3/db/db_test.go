package db_test

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
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

// test FindSet
func TestFindSet(t *testing.T) {
	memdb, err := mem.NewMem()
	goadapt.Tassert(t, err == nil, "NewMemDb() failed: %v", err)
	tx := memdb.NewTx(true)

	// add several intervals
	i0800_0900 := db.Tadd(tx, 5, "2024-01-01T08:00:00", "2024-01-01T09:00:00", 1.0)
	i0900_1000 := db.Tadd(tx, 10, "2024-01-01T09:00:00", "2024-01-01T10:00:00", 2.0)
	i1000_1100 := db.Tadd(tx, 20, "2024-01-01T10:00:00", "2024-01-01T11:00:00", 3.0)
	i1100_1200 := db.Tadd(tx, 30, "2024-01-01T11:00:00", "2024-01-01T12:00:00", 2.0)
	i1200_1300 := db.Tadd(tx, 40, "2024-01-01T12:00:00", "2024-01-01T12:45:00", 1.0)
	i1300_1400 := db.Tadd(tx, 50, "2024-01-01T13:00:00", "2024-01-01T14:00:00", 1.0)
	i1400_1500 := db.Tadd(tx, 60, "2024-01-01T14:00:00", "2024-01-01T15:00:00", 1.0)
	_ = i0800_0900
	_ = i0900_1000
	_ = i1000_1100
	_ = i1100_1200
	_ = i1200_1300
	_ = i1300_1400
	_ = i1400_1500

	// create a free interval that spans the gap between 12:45 and
	// 13:00 but don't add it to the db
	freeStart := time.Date(2024, 1, 1, 12, 45, 0, 0, time.UTC)
	freeEnd := time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)
	i1245_1300 := interval.NewInterval(0, freeStart, freeEnd, 0.0)

	// find the first set of intervals that are within a time range,
	// have a priority less than or equal to 1.0, and have a total
	// duration of at least 90 minutes
	start, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T08:00:00")
	goadapt.Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T15:00:00")
	goadapt.Ck(err)
	ivs, err := db.FindSet(tx, true, start, end, 90*time.Minute, 1.0)
	goadapt.Tassert(t, err == nil, "FindSet() failed: %v", err)
	goadapt.Tassert(t, len(ivs) == 3, "FindSet() failed: expected 3 intervals, got %v", spew.Sdump(ivs))
	goadapt.Tassert(t, i1200_1300.Equal(ivs[0]), "FindSet() failed: expected interval %v, got %v", i1200_1300, ivs[0])
	goadapt.Tassert(t, ivs[0].Priority == 1.0, "FindSet() failed: expected priority == 1.0, got %f", ivs[0].Priority)
	goadapt.Tassert(t, i1245_1300.Equal(ivs[1]), "FindSet() failed: expected interval %v, got %v", i1245_1300, ivs[1])
	goadapt.Tassert(t, ivs[1].Priority == 0.0, "FindSet() failed: expected priority == 0.0, got %f", ivs[1].Priority)
	goadapt.Tassert(t, i1300_1400.Equal(ivs[2]), "FindSet() failed: expected interval %v, got %v", i1300_1400, ivs[2])
	goadapt.Tassert(t, ivs[2].Priority == 1.0, "FindSet() failed: expected priority == 1.0, got %f", ivs[2].Priority)

	// find the last set of intervals that are within a time range,
	// have a priority less than or equal to 1.0, and have a total
	// duration of at least 90 minutes
	ivs, err = db.FindSet(tx, false, start, end, 90*time.Minute, 1.0)
	goadapt.Tassert(t, err == nil, "FindSet() failed: %v", err)
	// Pf("ivs: %v\n", ivs)
	goadapt.Tassert(t, len(ivs) == 2, "FindSet() failed: expected 2 intervals, got %v", spew.Sdump(ivs))
	goadapt.Tassert(t, i1400_1500.Equal(ivs[0]), "FindSet() failed: expected interval %v, got %v", i1400_1500, ivs[0])
	goadapt.Tassert(t, i1300_1400.Equal(ivs[1]), "FindSet() failed: expected interval %v, got %v", i1300_1400, ivs[1])

}

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
