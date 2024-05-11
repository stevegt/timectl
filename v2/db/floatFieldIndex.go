package db

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

// FloatFieldIndex is used to extract a float field from an object using
// reflection and builds an index on that field.
type FloatFieldIndex struct {
	Field string
}

// FromObject satisfies the go-memdb SingleIndexer interface.
func (f *FloatFieldIndex) FromObject(obj interface{}) (bool, []byte, error) {
	v := reflect.ValueOf(obj)
	v = reflect.Indirect(v) // Dereference the pointer if any

	fv := v.FieldByName(f.Field)
	if !fv.IsValid() {
		return false, nil,
			fmt.Errorf("field '%s' for %#v is invalid", f.Field, obj)
	}

	buf, err := encodeFloat(fv)
	if err != nil {
		return false, nil, err
	}

	return true, buf, nil
}

// FromArgs satisfies the go-memdb Indexer interface.
func (f *FloatFieldIndex) FromArgs(args ...interface{}) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("must provide only a single argument")
	}

	v := reflect.ValueOf(args[0])
	if !v.IsValid() {
		return nil, fmt.Errorf("%#v is invalid", args[0])
	}

	buf, err := encodeFloat(v)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func encodeFloat(v reflect.Value) (buf []byte, err error) {
	switch v.Kind() {
	case reflect.Float32:
		buf = make([]byte, 4)
		binary.BigEndian.PutUint32(buf[:], math.Float32bits(float32(v.Float())))
	case reflect.Float64:
		buf = make([]byte, 8)
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(v.Float()))
	default:
		return nil, fmt.Errorf("arg is of type %v; want a float", v.Kind())
	}
	return buf, nil
}
