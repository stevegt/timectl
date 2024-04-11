package interval

import "time"

// Tree represents a node in an interval tree.
type Tree struct {
    interval *Interval // The interval represented by this node
    left     *Tree     // Pointer to the left child
    right    *Tree     // Pointer to the right child
}

// NewTree creates and returns a new Tree node without an interval.
func NewTree() *Tree {
    return &Tree{}
}

// Insert adds a new interval to the tree, adjusting the structure as necessary.
func (t *Tree) Insert(newInterval *Interval) {
    if t.interval == nil && t.left == nil && t.right == nil {
        t.interval = newInterval
        return
    }

    if t.interval != nil && (t.left == nil && t.right == nil) {
        // The current node is a leaf node with an interval, split it.
        if newInterval.Start().Before(t.interval.Start()) || (newInterval.Start().Equal(t.interval.Start()) && newInterval.End().Before(t.interval.End())) {
            t.left = &Tree{interval: newInterval}
        } else {
            t.left = &Tree{interval: t.interval}
            t.right = &Tree{interval: newInterval}
            t.interval = nil
        }
        t.updateSpanningInterval()
        return
    }

    // Decide where to insert the new interval in the subtree.
    if newInterval.Start().Before(t.interval.Start()) || (newInterval.Start().Equal(t.interval.Start()) && newInterval.End().Before(t.interval.End())) {
        if t.left == nil {
            t.left = &Tree{}
        }
        t.left.Insert(newInterval)
    } else {
        if t.right == nil {
            t.right = &Tree{}
        }
        t.right.Insert(newInterval)
    }
    t.updateSpanningInterval()
}

// Conflicts finds and returns intervals in the tree that overlap with the given interval.
func (t *Tree) Conflicts(interval *Interval) []*Interval {
    var conflicts []*Interval
    if t.interval != nil && t.interval.Conflicts(interval) {
        conflicts = append(conflicts, t.interval)
    }
    if t.left != nil {
        conflicts = append(conflicts, t.left.Conflicts(interval)...)
    }
    if t.right != nil {
        conflicts = append(conflicts, t.right.Conflicts(interval)...)
    }
    return conflicts
}

// updateSpanningInterval updates the interval for the node to span its children,
// creating a new interval that covers both child intervals when necessary.
func (t *Tree) updateSpanningInterval() {
    var minStart, maxEnd time.Time
    var first = true

    if t.left != nil && t.left.interval != nil {
        if first {
            minStart, maxEnd = t.left.interval.Start(), t.left.interval.End()
            first = false
        } else {
            minStart = minTime(minStart, t.left.interval.Start())
            maxEnd = maxTime(maxEnd, t.left.interval.End())
        }
    }

    if t.right != nil && t.right.interval != nil {
        if first {
            minStart, maxEnd = t.right.interval.Start(), t.right.interval.End()
        } else {
            minStart = minTime(minStart, t.right.interval.Start())
            maxEnd = maxTime(maxEnd, t.right.interval.End())
        }
    }

    if !first { // if not the first run (i.e., if there was any child)
        t.interval = NewInterval(minStart, maxEnd)
    }
}

// minTime returns the earlier of two time.Time values.
func minTime(a, b time.Time) time.Time {
    if a.Before(b) {
        return a
    }
    return b
}

// maxTime returns the later of two time.Time values.
func maxTime(a, b time.Time) time.Time {
    if a.After(b) {
        return a
    }
    return b
}
