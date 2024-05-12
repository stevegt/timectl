package mem

import (
	"testing"
	"time"

	"github.com/stevegt/timectl/v3/db"

	"github.com/davecgh/go-spew/spew"
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/v3/interval"
)

func TestMemDb(t *testing.T) {
	// test go-memdb implementation of Db interface

	// open a new memdb
	memdb, err := NewMem()
	Tassert(t, err == nil, "NewMemDb() failed: %v", err)

	// get a write transaction
	tx := memdb.NewTx(true)

	// test Add
	start, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:00:00")
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:00:00")
	Ck(err)
	expect := interval.NewInterval(1, start, end, 1.0)
	err = tx.Add(expect)
	Tassert(t, err == nil, "Add() failed: %v", err)

	// test Get
	ivs, err := tx.FindFwd(start, end, 99.0)
	Tassert(t, err == nil, "Get() failed: %v", err)
	Tassert(t, len(ivs) == 1, "Get() failed: expected 1 interval, got %v", len(ivs))
	got := ivs[0]
	Tassert(t, expect.Equal(got), "Get() failed: expected interval %v, got %v", expect, got)
	Tassert(t, expect.Priority == got.Priority, "Get() failed: expected priority %f, got %f", expect.Priority, got.Priority)
}

func TestMemDbFind(t *testing.T) {
	// open a new memdb
	memdb, err := NewMem()
	Tassert(t, err == nil, "NewMemDb() failed: %v", err)

	// get a write transaction
	tx := memdb.NewTx(true)

	// add several intervals
	i0800_0900 := db.Tadd(tx, 5, "2024-01-01T08:00:00", "2024-01-01T09:00:00", 1.0)
	i0900_1000 := db.Tadd(tx, 10, "2024-01-01T09:00:00", "2024-01-01T10:00:00", 2.0)
	i1000_1100 := db.Tadd(tx, 20, "2024-01-01T10:00:00", "2024-01-01T11:00:00", 3.0)
	i1100_1200 := db.Tadd(tx, 30, "2024-01-01T11:00:00", "2024-01-01T12:00:00", 2.0)
	i1200_1300 := db.Tadd(tx, 40, "2024-01-01T12:00:00", "2024-01-01T13:00:00", 1.0)
	i1300_1400 := db.Tadd(tx, 50, "2024-01-01T13:00:00", "2024-01-01T14:00:00", 1.0)
	_ = i0800_0900
	_ = i1100_1200
	_ = i1200_1300
	_ = i1300_1400

	// get three intervals that overlap a time range
	start, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T09:30:00")
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T11:30:00")
	Ck(err)
	ivs, err := tx.FindFwd(start, end, 99.0)
	Tassert(t, err == nil, "Find() failed: %v", err)
	Tassert(t, len(ivs) == 3, "Find() failed: expected 3 intervals, got %v", spew.Sdump(ivs))
	Tassert(t, i0900_1000.Equal(ivs[0]), "Find() failed: expected interval %v, got %v", i0900_1000, ivs[0])
	Tassert(t, i1000_1100.Equal(ivs[1]), "Find() failed: expected interval %v, got %v", i1000_1100, ivs[1])
	Tassert(t, i1100_1200.Equal(ivs[2]), "Find() failed: expected interval %v, got %v", i1100_1200, ivs[2])

	// now try it again with a lower max priority
	ivs, err = tx.FindFwd(start, end, 2.0)
	Tassert(t, err == nil, "Find() failed: %v", err)
	Tassert(t, len(ivs) == 2, "Find() failed: expected 2 intervals, got %v", spew.Sdump(ivs))
	Tassert(t, i0900_1000.Equal(ivs[0]), "Find() failed: expected interval %v, got %v", i0900_1000, ivs[0])
	Tassert(t, i1100_1200.Equal(ivs[1]), "Find() failed: expected interval %v, got %v", i1100_1200, ivs[1])

}

// XXX test payload preservation
