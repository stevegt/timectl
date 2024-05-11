package storage

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

func TestMemDb(t *testing.T) {
	// test go-memdb implementation of Db interface

	// open a new memdb
	d, err := NewMem()
	Tassert(t, err == nil, "NewMemDb() failed: %v", err)

	// get a write transaction
	tx := d.NewTx(true)

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
	d, err := NewMem()
	Tassert(t, err == nil, "NewMemDb() failed: %v", err)

	// get a write transaction
	tx := d.NewTx(true)

	// add several intervals
	i0800_0900 := Tadd(tx, 5, "2024-01-01T08:00:00", "2024-01-01T09:00:00", 1.0)
	i0900_1000 := Tadd(tx, 10, "2024-01-01T09:00:00", "2024-01-01T10:00:00", 2.0)
	i1000_1100 := Tadd(tx, 20, "2024-01-01T10:00:00", "2024-01-01T11:00:00", 3.0)
	i1100_1200 := Tadd(tx, 30, "2024-01-01T11:00:00", "2024-01-01T12:00:00", 2.0)
	i1200_1300 := Tadd(tx, 40, "2024-01-01T12:00:00", "2024-01-01T13:00:00", 1.0)
	i1300_1400 := Tadd(tx, 50, "2024-01-01T13:00:00", "2024-01-01T14:00:00", 1.0)
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

// test FindSet
func TestMemDbFindSet(t *testing.T) {
	d, err := NewMem()
	Tassert(t, err == nil, "NewMemDb() failed: %v", err)
	tx := d.NewTx(true)

	// add several intervals
	i0800_0900 := Tadd(tx, 5, "2024-01-01T08:00:00", "2024-01-01T09:00:00", 1.0)
	i0900_1000 := Tadd(tx, 10, "2024-01-01T09:00:00", "2024-01-01T10:00:00", 2.0)
	i1000_1100 := Tadd(tx, 20, "2024-01-01T10:00:00", "2024-01-01T11:00:00", 3.0)
	i1100_1200 := Tadd(tx, 30, "2024-01-01T11:00:00", "2024-01-01T12:00:00", 2.0)
	i1200_1300 := Tadd(tx, 40, "2024-01-01T12:00:00", "2024-01-01T12:45:00", 1.0)
	i1300_1400 := Tadd(tx, 50, "2024-01-01T13:00:00", "2024-01-01T14:00:00", 1.0)
	i1400_1500 := Tadd(tx, 60, "2024-01-01T14:00:00", "2024-01-01T15:00:00", 1.0)
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
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T15:00:00")
	Ck(err)
	ivs, err := tx.FindSet(true, start, end, 90*time.Minute, 1.0)
	Tassert(t, err == nil, "FindSet() failed: %v", err)
	Tassert(t, len(ivs) == 3, "FindSet() failed: expected 3 intervals, got %v", spew.Sdump(ivs))
	Tassert(t, i1200_1300.Equal(ivs[0]), "FindSet() failed: expected interval %v, got %v", i1200_1300, ivs[0])
	Tassert(t, ivs[0].Priority == 1.0, "FindSet() failed: expected priority == 1.0, got %f", ivs[0].Priority)
	Tassert(t, i1245_1300.Equal(ivs[1]), "FindSet() failed: expected interval %v, got %v", i1245_1300, ivs[1])
	Tassert(t, ivs[1].Priority == 0.0, "FindSet() failed: expected priority == 0.0, got %f", ivs[1].Priority)
	Tassert(t, i1300_1400.Equal(ivs[2]), "FindSet() failed: expected interval %v, got %v", i1300_1400, ivs[2])
	Tassert(t, ivs[2].Priority == 1.0, "FindSet() failed: expected priority == 1.0, got %f", ivs[2].Priority)

	// find the last set of intervals that are within a time range,
	// have a priority less than or equal to 1.0, and have a total
	// duration of at least 90 minutes
	ivs, err = tx.FindSet(false, start, end, 90*time.Minute, 1.0)
	Tassert(t, err == nil, "FindSet() failed: %v", err)
	// Pf("ivs: %v\n", ivs)
	Tassert(t, len(ivs) == 2, "FindSet() failed: expected 2 intervals, got %v", spew.Sdump(ivs))
	Tassert(t, i1400_1500.Equal(ivs[0]), "FindSet() failed: expected interval %v, got %v", i1400_1500, ivs[0])
	Tassert(t, i1300_1400.Equal(ivs[1]), "FindSet() failed: expected interval %v, got %v", i1300_1400, ivs[1])

}

// XXX test payload preservation
