package tree

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
	"github.com/stevegt/timectl/interval"
)

// Get is a test helper function that returns the tree node at the
// given path in the tree.  pathStr is the path to the node in the
// tree, where 'l' means to go left and 'r' means to go right.  An
// empty pathStr means to return the root node.
func Get(tree *Node, pathStr string) *Node {
	path := []rune(pathStr)
	if len(path) == 0 {
		return tree
	}
	switch path[0] {
	case 'l':
		Assert(tree.Left() != nil, "No left node")
		return Get(tree.Left(), string(path[1:]))
	case 'r':
		Assert(tree.Right() != nil, "No right node")
		return Get(tree.Right(), string(path[1:]))
	default:
		Assert(false, "Invalid path %v", pathStr)
	}
	return nil
}

// Expect is a test helper function that checks if the given tree
// node's interval has the expected start and end times and priority.
// pathStr is the path to the node in the tree, where 'l'
// means to go left and 'r' means to go right.  An empty pathStr means
// to check the root node.
func Expect(tree *Node, pathStr, startStr, endStr string, priority float64) error {
	expectStr := Spf("expect %v %v %v %v\n", pathStr, startStr, endStr, priority)
	node := Get(tree, pathStr)
	if node == nil {
		return fmt.Errorf("no node at path: %v", pathStr)
	}
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	ev := interval.NewInterval(start, end, priority)
	ok := node.Interval().Equal(ev)
	ok = ok && node.Interval().Priority() == priority
	if !ok {
		return fmt.Errorf("%vExpected %v, got %v", expectStr, ev, node.Interval())
	}
	return nil
}

// InsertExpect is a test helper function that inserts an interval
// into the tree and checks if the tree has the expected structure.
func InsertExpect(tree *Node, pathStr, startStr, endStr string, priority float64) error {
	interval := Insert(tree, startStr, endStr, priority)
	if interval == nil {
		return fmt.Errorf("Failed to insert interval")
	}
	return Expect(tree, pathStr, startStr, endStr, priority)
}

// NewInterval is a test helper function that creates a new interval
// with the given start and end times and priority.
func NewInterval(startStr, endStr string, priority float64) interval.Interval {
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	return interval.NewInterval(start, end, priority)
}

// Insert is a test helper function that inserts an interval into the
// tree and returns the interval that was inserted.
func Insert(tree *Node, startStr, endStr string, priority float64) interval.Interval {
	interval := NewInterval(startStr, endStr, priority)
	// Insert adds a new interval to the tree, adjusting the structure as
	// necessary.  Insertion fails if the new interval conflicts with any
	// existing interval in the tree.
	// Pf("Inserting interval: %v\n", interval)
	ok := tree.Insert(interval)
	if !ok {
		return nil
	}
	return interval
}

// Match is a test helper function that checks if the given interval
// has the expected start and end times and priority.
func Match(iv interval.Interval, startStr, endStr string, priority float64) error {
	start, err := time.Parse(time.RFC3339, startStr)
	Ck(err)
	end, err := time.Parse(time.RFC3339, endStr)
	Ck(err)
	ev := interval.NewInterval(start, end, priority)
	if !iv.Equal(ev) {
		return fmt.Errorf("Expected %v, got %v", ev, iv)
	}
	return nil
}

// SaveDot is a test helper function that saves the tree as a dot file
func SaveDot(tree *Node) {
	// get caller's file and line number
	_, file, line, ok := runtime.Caller(1)
	Assert(ok, "Failed to get caller")
	// keep only the file name, throw away the path
	_, file = filepath.Split(file)
	fn := fmt.Sprintf("/tmp/%s:%d.dot", file, line)
	buf := []byte(tree.AsDot(nil))
	err := ioutil.WriteFile(fn, buf, 0644)
	Ck(err)
}

// Verify is a test helper function that verifies the tree.  If
// there is an error, it shows the tree as a dot file.
func Verify(t *testing.T, tree *Node, ckBalance bool, show bool) {
	err := tree.Verify(false)
	if err != nil {
		// get caller's file and line number
		_, file, line, ok := runtime.Caller(1)
		Assert(ok, "Failed to get caller")
		msg := Spf("%v:%v %v\n", file, line, err)
		Pl(msg)
		if show {
			ShowDot(tree, false)
		}
		t.Fatal(msg)
	}
}

// Dump is a helper function that prints the tree structure to
// stdout.
func Dump(tree *Node, path string) {
	// fmt.Printf("maxGap: %v interval: %v\n", tree.maxGap, tree.Interval())
	if tree.Left() != nil {
		Dump(tree.Left(), path+"l")
	}
	fmt.Printf("%-10v: %v\n", path, tree.Interval())
	if tree.Right() != nil {
		Dump(tree.Right(), path+"r")
	}
}

// ShowDot displays the tree in xdot.  If bg is true, then the xdot
// window is displayed from a background process.
func ShowDot(tree *Node, bg bool) {
	dot := tree.AsDot(nil)
	// call 'xdot -' passing the dot file as input
	cmd := exec.Command("xdot", "-")
	cmd.Stdin = strings.NewReader(dot)
	if bg {
		cmd.Start()
		go cmd.Wait()
		return
	}
	cmd.Run()
}
