package store

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"time"
)

// TimeFieldIndex is an index that indexes time.Time fields.
type TimeFieldIndex struct {
	Field string
}

// FromObject satisfies the go-memdb SingleIndexer interface.
func (i *TimeFieldIndex) FromObject(obj interface{}) (bool, []byte, error) {
	v := reflect.ValueOf(obj)
	v = reflect.Indirect(v) // Dereference the pointer if any

	fv := v.FieldByName(i.Field)
	if !fv.IsValid() {
		return false, nil,
			fmt.Errorf("field '%s' for %#v is invalid", i.Field, obj)
	}

	buf, err := encodeTime(fv)
	if err != nil {
		return false, nil, err
	}

	return true, buf, nil
}

// FromArgs satisfies the go-memdb Indexer interface.
func (i *TimeFieldIndex) FromArgs(args ...interface{}) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("must provide only a single argument")
	}

	v := reflect.ValueOf(args[0])
	if !v.IsValid() {
		return nil, fmt.Errorf("%#v is invalid", args[0])
	}

	buf, err := encodeTime(v)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func encodeTime(v reflect.Value) (buf []byte, err error) {
	// Check if the field is a time.Time
	timeType := reflect.TypeOf(time.Time{})
	if v.Type() != timeType {
		return nil, fmt.Errorf("field is not a time.Time")
	}

	// Dereference pointers and interfaces
	k := v.Kind()
	var val reflect.Value
	if k == reflect.Ptr || k == reflect.Interface {
		if v.IsNil() {
			return nil, fmt.Errorf("time.Time pointer is nil")
		}
		val = v.Elem()
	} else {
		val = v
	}

	// convert time.Time to int64 nanoseconds
	nanoInt64 := val.Interface().(time.Time).UnixNano()

	// convert int64 nanoseconds to uint64
	size := 8
	// from go-memdb's index.go:
	// This bit flips the sign bit on any sized signed twos-complement integer,
	// which when truncated to a uint of the same size will bias the value such
	// that the maximum negative int becomes 0, and the maximum positive int
	// becomes the maximum positive uint.
	scaled := nanoInt64 ^ int64(-1<<(size*8-1))

	nanoUint64 := uint64(scaled)

	// encode uint64 to []byte buffer
	buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, nanoUint64)

	return buf, nil
}
