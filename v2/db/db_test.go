package db

import (
	"testing"
	"time"

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
	gots, err := tx.Find(start, end, 99.0)
	Tassert(t, err == nil, "Get() failed: %v", err)
	Tassert(t, len(gots) == 1, "Get() failed: expected 1 interval, got %d", len(gots))
	got := gots[0]
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
	i0900_1000 := Tadd(tx, 10, "2024-01-01T09:00:00", "2024-01-01T10:00:00", 1.0)
	i1000_1100 := Tadd(tx, 20, "2024-01-01T10:00:00", "2024-01-01T11:00:00", 1.0)
	i1100_1200 := Tadd(tx, 30, "2024-01-01T11:00:00", "2024-01-01T12:00:00", 1.0)
	i1200_1300 := Tadd(tx, 40, "2024-01-01T12:00:00", "2024-01-01T13:00:00", 1.0)
	i1300_1400 := Tadd(tx, 50, "2024-01-01T13:00:00", "2024-01-01T14:00:00", 1.0)
	_ = i0800_0900
	_ = i1100_1200
	_ = i1200_1300
	_ = i1300_1400

	// get two intervals that overlap a time range
	start, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T09:30:00")
	Ck(err)
	end, err := time.Parse("2006-01-02T15:04:05", "2024-01-01T10:30:00")
	Ck(err)
	gots, err := tx.Find(start, end, 99.0)
	Tassert(t, err == nil, "Find() failed: %v", err)
	Tassert(t, len(gots) == 2, "Find() failed: expected 2 intervals, got %d", len(gots))
	Tassert(t, i0900_1000.Equal(gots[0]), "Find() failed: expected interval %v, got %v", i0900_1000, gots[0])
	Tassert(t, i1000_1100.Equal(gots[1]), "Find() failed: expected interval %v, got %v", i1000_1100, gots[1])

}

// test FindSet
// _, err = tx.FindSet(start, end, 99.0)
// Tassert(t, err == nil, "FindSet() failed: %v", err)
