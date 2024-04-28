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
