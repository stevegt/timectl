package db

import (
	"time"

	"github.com/stevegt/timectl/interval"
)

// Db is an interface for an interval data storage system.  It
// abstracts the underlying storage system.
type Db interface {
	// Close closes the database by saving any remaining changes (if
	// the database is persistent) and releasing any resources.
	Close() error

	// NewTx returns a transaction for the database.  If the write
	// parameter is true, the transaction is a write transaction.
	NewTx(write bool) Tx
}

// Tx is an interface for a database transaction.  It provides
// methods for adding, changing, deleting, and querying intervals.
type Tx interface {

	// Commit commits the transaction.  If the transaction is a write
	// transaction, it writes the changes to the database.
	Commit()

	// Abort aborts the transaction.  If the transaction is a write
	// transaction, it discards the changes.  If the transaction is a
	// read transaction, it releases the resources.
	Abort()

	// Add adds an interval to the database.  If the interval
	// conflicts with an existing interval, it returns an error.
	Add(iv *interval.Interval) error

	// SetPriority sets the priority of an interval in the database.  If
	// the interval does not exist, it returns an error.
	// SetPriority(iv interval.Interval, priority float64) error

	// Delete deletes an interval from the database.  If the
	// interval does not exist, it returns an error.
	Delete(iv *interval.Interval) error

	// FindFwd returns all intervals that intersect with the given
	// given start and end time and are lower than the given priority.
	// The results are ordered by ascending end time. The results
	// include synthetic free intervals that represent the time slots
	// between the intervals.
	FindFwd(minStart, maxEnd time.Time, maxPriority float64) ([]*interval.Interval, error)
	FindFwdIter(minStart, maxEnd time.Time, maxPriority float64) (Iterator, error)

	// FindRev is the same as FindFwd, but the results are ordered by
	// descending start time.
	FindRev(minStart, maxEnd time.Time, maxPriority float64) ([]*interval.Interval, error)
	FindRevIter(minStart, maxEnd time.Time, maxPriority float64) (Iterator, error)

	// FindSet returns a contiguous set of intervals that intersect
	// with the given start and end time, are lower than the given
	// priority, and whose total duration is greater than or equal to the
	// given duration.  The first parameter indicates whether the set
	// should be the first or last match found within the given time
	// range.
	// The results include synthetic free intervals that represent the
	// time slots between the intervals.
	FindSet(first bool, minStart, maxEnd time.Time, minDuration time.Duration, maxPriority float64) ([]*interval.Interval, error)

	// IterateDown returns an iterator that iterates over all intervals
	// in the database in descending order of priority.
	// IteratorDown() Iterator

	// IterateForward returns an iterator that iterates over all intervals
	// in the database in ascending order of start time.
	// IteratorForward() Iterator
}

// Iterator is an interface for iterating over intervals in a database.
type Iterator interface {
	// Next returns the next interval in the iteration.  If there are no
	// more intervals, it returns nil.
	Next() *interval.Interval
}
