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

    if t.interval != nil {
        // Decide to insert left or right based on the start time
        if newInterval.Start().Before(t.interval.Start()) {
            if t.left == nil {
                t.left = &Tree{interval: newInterval}
            } else {
                t.left.Insert(newInterval)
            }
        } else {
            if t.right == nil {
                t.right = &Tree{interval: newInterval}
            } else {
                t.right.Insert(newInterval)
            }
        }
    }

    // After inserting, update the interval of the current node to ensure it spans its children.
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
    if t.left != nil || t.right != nil {
        var minStart, maxEnd time.Time = t.interval.Start(), t.interval.End()

        if t.left != nil {
            minStart = minTime(minStart, t.left.interval.Start())
            maxEnd = maxTime(maxEnd, t.left.interval.End())
        }
        if t.right != nil {
            minStart = minTime(minStart, t.right.interval.Start())
            maxEnd = maxTime(maxEnd, t.right.interval.End())
        }
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
