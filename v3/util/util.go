package util

import (
	"time"
	// . "github.com/stevegt/goadapt"
)

// MinTime returns the earlier of two time.Time values.
func MinTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// MaxTime returns the later of two time.Time values.
func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// MaxDuration returns the longer of two time.Duration values.
func MaxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// AbsDuration returns the absolute value of a time.Duration.
func AbsDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// OnOrBefore returns true if the first time is on or before the second time.
func OnOrBefore(a, b time.Time) bool {
	return !a.After(b)
}

// OnOrAfter returns true if the first time is on or after the second time.
func OnOrAfter(a, b time.Time) bool {
	return !a.Before(b)
}
