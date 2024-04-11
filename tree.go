package interval

// TreeNode represents a node in the interval tree.
type TreeNode struct {
    Interval *Interval // The interval stored in this node
    Left     *TreeNode // Pointer to the left child
    Right    *TreeNode // Pointer to the right child
}

// Tree represents an interval tree.
type Tree struct {
    Root *TreeNode // Root node of the tree
}

// NewTree creates and returns a new Tree.
func NewTree() *Tree {
    return &Tree{}
}

// insertNode is a recursive helper function for inserting a new node into the tree.
func insertNode(node, newNode *TreeNode) *TreeNode {
    if node == nil {
        return newNode
    }
    // This is a simplified insertion based on start time only, for demonstration purposes.
    if newNode.Interval.Start().Before(node.Interval.Start()) {
        node.Left = insertNode(node.Left, newNode)
    } else {
        node.Right = insertNode(node.Right, newNode)
    }
    return node
}

// Insert adds a new interval to the tree.
func (t *Tree) Insert(interval *Interval) {
    newNode := &TreeNode{Interval: interval}
    if t.Root == nil {
        t.Root = newNode
    } else {
        t.Root = insertNode(t.Root, newNode)
    }
}

// Conflicts checks if the given interval conflicts with any interval in the tree.
// It returns a slice of intervals that conflict.
func (t *Tree) Conflicts(interval *Interval) []*Interval {
    var conflicts []*Interval
    var check func(node *TreeNode)
    check = func(node *TreeNode) {
        if node == nil {
            return
        }
        if node.Interval.Conflicts(interval) {
            conflicts = append(conflicts, node.Interval)
        }
        check(node.Left)
        check(node.Right)
    }
    check(t.Root)
    return conflicts
}
