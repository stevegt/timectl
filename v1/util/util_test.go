package util

import (
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
)

// test mintime
func TestMinTime(t *testing.T) {
	t1Str := "2017-01-01 00:00:00"
	t2Str := "2017-01-01 00:00:01"

	t1, err := time.Parse("2006-01-02 15:04:05", t1Str)
	Tassert(t, err == nil, "time parse error")
	t2, err := time.Parse("2006-01-02 15:04:05", t2Str)
	Tassert(t, err == nil, "time parse error")

	Tassert(t, MinTime(t1, t2) == t1, "MinTime failed")
	Tassert(t, MinTime(t2, t1) == t1, "MinTime failed")
}

// test maxtime
func TestMaxTime(t *testing.T) {
	t1Str := "2017-01-01 00:00:00"
	t2Str := "2017-01-01 00:00:01"

	t1, err := time.Parse("2006-01-02 15:04:05", t1Str)
	Tassert(t, err == nil, "time parse error")
	t2, err := time.Parse("2006-01-02 15:04:05", t2Str)
	Tassert(t, err == nil, "time parse error")

	Tassert(t, MaxTime(t1, t2) == t2, "MaxTime failed")
	Tassert(t, MaxTime(t2, t1) == t2, "MaxTime failed")
}

// test maxduration
func TestMaxDuration(t *testing.T) {
	d1, err := time.ParseDuration("1s")
	Tassert(t, err == nil, "duration parse error")
	d2, err := time.ParseDuration("2s")
	Tassert(t, err == nil, "duration parse error")

	Tassert(t, MaxDuration(d1, d2) == d2, "MaxDuration failed")
	Tassert(t, MaxDuration(d2, d1) == d2, "MaxDuration failed")
}
