package datastructure

import (
	"errors"
	"fmt"
)

const (
	DEFAULT_ORDER = 4  // Default order of the B+ tree
	MIN_ORDER     = 3  // Minimum allowed order
	MAX_ORDER     = 20 // Maximum allowed order
)

type Node struct {
	leaf     bool
	keys     []int
	children []*Node
	next     *Node // Used for leaf nodes to link to the next leaf
}

type BPlusTree struct {
	root  *Node
	order int // Order of the B+ tree
}

func NewBPlusTree(order int) (*BPlusTree, error) {
	if order < MIN_ORDER || order > MAX_ORDER {
		return nil, errors.New("order must be between MIN_ORDER and MAX_ORDER")
	}
	return &BPlusTree{root: nil, order: order}, nil
}

// Insert a key into the B+ tree
func (t *BPlusTree) Insert(key int) {
	if t.root == nil {
		t.root = &Node{
			leaf: true,
			keys: []int{key},
		}
		return
	}

	if len(t.root.keys) == t.order-1 {
		newRoot := &Node{
			leaf:     false,
			children: []*Node{t.root},
		}
		t.splitChild(newRoot, 0)
		t.root = newRoot
	}

	t.insertNonFull(t.root, key)
}

// Insert into a non-full node
func (t *BPlusTree) insertNonFull(node *Node, key int) {
	i := len(node.keys) - 1

	if node.leaf {
		// Insert key into the leaf node
		for i >= 0 && key < node.keys[i] {
			i--
		}
		node.keys = append(node.keys, 0)
		copy(node.keys[i+2:], node.keys[i+1:])
		node.keys[i+1] = key
	} else {
		// Find the child to insert into
		for i >= 0 && key < node.keys[i] {
			i--
		}
		i++
		if len(node.children[i].keys) == t.order-1 {
			t.splitChild(node, i)
			if key > node.keys[i] {
				i++
			}
		}
		t.insertNonFull(node.children[i], key)
	}
}

// Split a child node
func (t *BPlusTree) splitChild(parent *Node, index int) {
	child := parent.children[index]
	newNode := &Node{
		leaf: child.leaf,
	}

	// Split keys
	mid := len(child.keys) / 2
	newNode.keys = make([]int, len(child.keys)-mid-1)
	copy(newNode.keys, child.keys[mid+1:])
	child.keys = child.keys[:mid]

	// Split children if not a leaf
	if !child.leaf {
		newNode.children = make([]*Node, len(child.children)-mid-1)
		copy(newNode.children, child.children[mid+1:])
		child.children = child.children[:mid+1]
	}

	// Insert the middle key into the parent
	parent.keys = append(parent.keys, 0)
	copy(parent.keys[index+1:], parent.keys[index:])
	parent.keys[index] = child.keys[mid]

	// Insert the new node into the parent's children
	parent.children = append(parent.children, nil)
	copy(parent.children[index+2:], parent.children[index+1:])
	parent.children[index+1] = newNode

	// Update leaf node links
	if child.leaf {
		newNode.next = child.next
		child.next = newNode
	}
}

// Delete a key from the B+ tree
func (t *BPlusTree) Delete(key int) {
	if t.root == nil {
		return
	}

	t.deleteKey(t.root, key)

	// If the root has no keys, make its first child the new root
	if len(t.root.keys) == 0 {
		if t.root.leaf {
			t.root = nil
		} else {
			t.root = t.root.children[0]
		}
	}
}

// Delete a key from a node
func (t *BPlusTree) deleteKey(node *Node, key int) {
	i := 0
	for i < len(node.keys) && key > node.keys[i] {
		i++
	}

	if node.leaf {
		// Key is in this leaf node
		if i < len(node.keys) && key == node.keys[i] {
			node.keys = append(node.keys[:i], node.keys[i+1:]...)
		}
	} else {
		// Key is in a child node
		t.deleteKey(node.children[i], key)

		// Rebalance the tree if necessary
		if len(node.children[i].keys) < (t.order-1)/2 {
			t.rebalance(node, i)
		}
	}
}

// Rebalance the tree after deletion
func (t *BPlusTree) rebalance(parent *Node, index int) {
	if index > 0 && len(parent.children[index-1].keys) > (t.order-1)/2 {
		// Borrow from the left sibling
		t.rotateRight(parent, index-1)
	} else if index < len(parent.children)-1 && len(parent.children[index+1].keys) > (t.order-1)/2 {
		// Borrow from the right sibling
		t.rotateLeft(parent, index)
	} else {
		// Merge with a sibling
		if index > 0 {
			t.mergeNodes(parent, index-1)
		} else {
			t.mergeNodes(parent, index)
		}
	}
}

// Rotate right (borrow from left sibling)
func (t *BPlusTree) rotateRight(parent *Node, index int) {
	child := parent.children[index]
	rightSibling := parent.children[index+1]

	// Move a key from the parent to the child
	child.keys = append(child.keys, parent.keys[index])
	parent.keys[index] = rightSibling.keys[0]
	rightSibling.keys = rightSibling.keys[1:]

	if !child.leaf {
		// Move the child pointer from the right sibling to the child
		child.children = append(child.children, rightSibling.children[0])
		rightSibling.children = rightSibling.children[1:]
	}
}

// Rotate left (borrow from right sibling)
func (t *BPlusTree) rotateLeft(parent *Node, index int) {
	child := parent.children[index]
	leftSibling := parent.children[index-1]

	// Move a key from the parent to the child
	child.keys = append([]int{parent.keys[index-1]}, child.keys...)
	parent.keys[index-1] = leftSibling.keys[len(leftSibling.keys)-1]
	leftSibling.keys = leftSibling.keys[:len(leftSibling.keys)-1]

	if !child.leaf {
		// Move the child pointer from the left sibling to the child
		child.children = append([]*Node{leftSibling.children[len(leftSibling.children)-1]}, child.children...)
		leftSibling.children = leftSibling.children[:len(leftSibling.children)-1]
	}
}

// Merge two nodes
func (t *BPlusTree) mergeNodes(parent *Node, index int) {
	child := parent.children[index]
	rightSibling := parent.children[index+1]

	// Move the key from the parent to the child
	child.keys = append(child.keys, parent.keys[index])
	child.keys = append(child.keys, rightSibling.keys...)

	if !child.leaf {
		// Move the children from the right sibling to the child
		child.children = append(child.children, rightSibling.children...)
	}

	// Remove the key and child from the parent
	parent.keys = append(parent.keys[:index], parent.keys[index+1:]...)
	parent.children = append(parent.children[:index+1], parent.children[index+2:]...)
}

// Search for a key in the B+ tree
func (t *BPlusTree) Search(key int) bool {
	return t.searchNode(t.root, key)
}

func (t *BPlusTree) searchNode(node *Node, key int) bool {
	if node == nil {
		return false
	}

	i := 0
	for i < len(node.keys) && key > node.keys[i] {
		i++
	}

	if i < len(node.keys) && key == node.keys[i] {
		return true
	}

	if node.leaf {
		return false
	}

	return t.searchNode(node.children[i], key)
}

// Traverse the B+ tree
func (t *BPlusTree) Traverse() {
	t.traverseNode(t.root)
}

func (t *BPlusTree) traverseNode(node *Node) {
	if node == nil {
		return
	}

	for i := 0; i < len(node.keys); i++ {
		fmt.Printf("%d ", node.keys[i])
	}

	if !node.leaf {
		for i := 0; i < len(node.children); i++ {
			t.traverseNode(node.children[i])
		}
	}
}

func main() {
	tree, err := NewBPlusTree(DEFAULT_ORDER)
	if err != nil {
		fmt.Println("Error creating B+ tree:", err)
		return
	}

	keys := []int{10, 20, 5, 6, 12, 30, 7, 17}

	for _, key := range keys {
		tree.Insert(key)
	}

	fmt.Println("B+ Tree traversal after insertion:")
	tree.Traverse()
	fmt.Println()

	fmt.Println("Search for 6:", tree.Search(6))
	fmt.Println("Search for 15:", tree.Search(15))

	tree.Delete(6)
	tree.Delete(12)

	fmt.Println("B+ Tree traversal after deletion:")
	tree.Traverse()
	fmt.Println()
}
