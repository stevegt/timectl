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

func TestMemDbFindSet(t *testing.T) {
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

	// test FindSet
	// _, err = tx.FindSet(start, end, 99.0)
	// Tassert(t, err == nil, "FindSet() failed: %v", err)
}
