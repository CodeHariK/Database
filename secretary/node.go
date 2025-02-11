package secretary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/binstruct"
)

/*
*
For a B+ tree of order  m :
*	Internal Nodes:
  - Have at most  m-1  keys (separators).
  - Have at most  m  children (pointers).
  - Therefore, KeyOffsets should have  m  elements (one for each child).

*	Leaf Nodes:
  - Have at most  m-1  keys (actual data keys).
  - Have exactly  m-1  record offsets (since each key maps to a record).
  - Therefore, KeyOffsets should have  m-1  elements (one per key).
*/
func (tree *BTree) NodeCheck(n *Node) (*Node, error) {
	if n.NumKeys > (tree.Order - 1) {
		return nil, ErrorNumKeysMoreThanOrder
	}

	// A node is either a leaf (has records) or an internal node (has children), not both
	if n.records != nil && n.children != nil {
		return nil, ErrorNodeIsEitherLeaforInternal
	}

	if n.children != nil {
		// Internal node
		if len(n.Keys) != int(n.NumKeys) ||
			len(n.KeyOffsets) != int(n.NumKeys+1) || // Should match children count
			len(n.children) != int(n.NumKeys+1) {
			return nil, ErrorNumKeysNotMatching
		}
	} else if n.records != nil {
		// Leaf node
		if len(n.Keys) != int(n.NumKeys) ||
			len(n.KeyOffsets) != int(n.NumKeys) || // Should match record count
			len(n.records) != int(n.NumKeys) {
			return nil, ErrorNumKeysNotMatching
		}
	}

	// Validate each key size and key offset
	for i, el := range n.Keys {
		if len(el) != KEY_SIZE {
			return nil, ErrorInvalidKey
		}

		// Bounds check for KeyOffsets
		if i < len(n.KeyOffsets) {
			if err := tree.dataLocationCheck(n.KeyOffsets[i]); err != nil {
				return nil, err
			}
		}
	}

	return n, nil
}

func (tree *BTree) NewNode(
	parentOffset, nextOffset, prevOffset DataLocation,
	numKeys int,
	keyOffsets []DataLocation,
	keys [][]byte,
) (*Node, error) {
	n := &Node{
		ParentOffset: parentOffset,
		NextOffset:   nextOffset,
		PrevOffset:   prevOffset,

		NumKeys:    uint8(numKeys),
		KeyOffsets: keyOffsets,
		Keys:       keys,
	}

	return tree.NodeCheck(n)
}

func (n *Node) IsLeaf() bool {
	return len(n.children) == 0
}

func (tree *BTree) saveRoot() error {
	rootHeader, err := binstruct.Serialize(*tree.root)
	if err != nil {
		return err
	}

	return tree.nodeBatchStore.WriteAt(SECRETARY_HEADER_LENGTH, rootHeader)
}

func (tree *BTree) dataLocationCheck(location DataLocation) error {
	if location == -1 {
		return ErrorInvalidDataLocation
	}
	return nil
}

//------------------------------------------------------------------

// Create a new internal node
func (tree *BTree) createInternalNode() *Node {
	return &Node{
		Keys:     make([][]byte, 0),
		children: make([]*Node, 0),
		NumKeys:  0,
	}
}

// Create a new leaf node
func (tree *BTree) createLeafNode() *Node {
	return &Node{
		Keys:    make([][]byte, 0),
		records: make([]*Record, 0),
		NumKeys: 0,
	}
}

// Find the appropriate leaf node
func (tree *BTree) findLeafNode(key []byte) *Node {
	current := tree.root
	for len(current.children) > 0 {
		i := sort.Search(len(current.Keys), func(i int) bool {
			return string(current.Keys[i]) > string(key)
		})
		current = current.children[i]
	}
	return current
}

// Insert key-value into a leaf node
func (tree *BTree) insertIntoLeaf(leaf *Node, key []byte, value []byte) {
	i := sort.Search(
		len(leaf.Keys),
		func(i int) bool {
			return string(leaf.Keys[i]) > string(key)
		},
	)

	leaf.Keys = append(
		leaf.Keys[:i],
		append([][]byte{key}, leaf.Keys[i:]...)...)

	leaf.records = append(
		leaf.records[:i],
		append([]*Record{
			{
				Key:   key,
				Value: value,
			},
		}, leaf.records[i:]...)...,
	)
	leaf.NumKeys++
}

// Split a leaf node
func (tree *BTree) splitLeaf(leaf *Node) {
	mid := len(leaf.Keys) / 2
	newLeaf := tree.createLeafNode()

	newLeaf.Keys = append(newLeaf.Keys, leaf.Keys[mid:]...)
	newLeaf.records = append(newLeaf.records, leaf.records[mid:]...)
	newLeaf.NumKeys = uint8(len(newLeaf.Keys))

	leaf.Keys = leaf.Keys[:mid]
	leaf.records = leaf.records[:mid]
	leaf.NumKeys = uint8(len(leaf.Keys))

	/**
	 * +++++++++++  +++++++++++
	 * +  leaf   +  +  right  +
	 * +++++++++++  +++++++++++
	 * +++++++++++  +++++++++++  +++++++++++
	 * +  leaf   +  + newLeaf +  +  right  +
	 * +++++++++++  +++++++++++  +++++++++++
	 */

	newLeaf.next = leaf.next
	if leaf.next != nil {
		leaf.next.prev = newLeaf
	}
	leaf.next = newLeaf
	newLeaf.prev = leaf

	tree.insertIntoParent(leaf, newLeaf.Keys[0], newLeaf)
}

// Insert into an internal node
func (tree *BTree) insertIntoParent(left *Node, key []byte, right *Node) {
	// If left is leaf and root, then create new root
	if left.parent == nil {
		newRoot := tree.createInternalNode()
		newRoot.Keys = [][]byte{key}
		newRoot.children = []*Node{left, right}
		left.parent = newRoot
		right.parent = newRoot
		tree.root = newRoot
		return
	}

	parent := left.parent
	insertIdx := sort.Search(
		len(parent.Keys),
		func(i int) bool {
			return string(parent.Keys[i]) > string(key)
		})

	parent.Keys = append(
		parent.Keys[:insertIdx],
		append(
			[][]byte{key},
			parent.Keys[insertIdx:]...,
		)...)

	parent.children = append(
		parent.children[:insertIdx+1],
		append(
			[]*Node{right},
			parent.children[insertIdx+1:]...,
		)...)

	right.parent = parent
	parent.NumKeys++

	fmt.Println("\ni ", parent.NumKeys, len(parent.Keys), int(tree.Order-1), "split:", int(parent.NumKeys) > int(tree.Order-1), utils.BytesToStrings(parent.Keys))

	if int(parent.NumKeys) > int(tree.Order-1) {
		tree.splitInternal(parent)
	}
}

// Split an internal node
func (tree *BTree) splitInternal(node *Node) {
	mid := len(node.Keys) / 2

	newInternal := tree.createInternalNode()
	newInternal.Keys = append(
		newInternal.Keys,
		node.Keys[mid+1:]...)
	newInternal.children = append(
		newInternal.children,
		node.children[mid+1:]...)
	newInternal.NumKeys = uint8(len(newInternal.Keys))

	for _, child := range newInternal.children {
		child.parent = newInternal
	}

	promotedKey := node.Keys[mid]
	node.Keys = node.Keys[:mid]
	node.children = node.children[:mid+1]
	node.NumKeys = uint8(len(node.Keys))

	tree.insertIntoParent(node, promotedKey, newInternal)
}

// Insert a key-value pair into the B+ Tree
func (tree *BTree) Insert(key []byte, value []byte) error {
	if len(key) != 16 {
		return ErrorInvalidKey
	}

	if tree.root == nil {
		tree.root = tree.createLeafNode()
		tree.insertIntoLeaf(tree.root, key, value)
		return nil
	}

	leaf := tree.findLeafNode(key)
	for _, k := range leaf.Keys {
		if string(k) == string(key) {
			return ErrorDuplicateKey
		}
	}

	tree.insertIntoLeaf(leaf, key, value)

	fmt.Println("\nl ", leaf.NumKeys, len(leaf.Keys), int(tree.Order-1), "split:", int(leaf.NumKeys) > int(tree.Order-1), utils.BytesToStrings(leaf.Keys))

	if int(leaf.NumKeys) > int(tree.Order-1) {
		tree.splitLeaf(leaf)
	}
	return nil
}

// Update a key-value pair in the B+ Tree
func (tree *BTree) Update(key []byte, value []byte) error {
	if len(key) != 16 {
		return ErrorInvalidKey
	}

	leaf := tree.findLeafNode(key)
	for i, k := range leaf.Keys {
		if string(k) == string(key) {
			leaf.records[i].Value = value
			return nil
		}
	}
	return ErrorKeyNotFound
}

// ------------------------------------------------------------------

// BulkLoad inserts sorted records into the B+ Tree efficiently
func (tree *BTree) BulkLoad(sortedRecords []*Record) {
	if len(sortedRecords) == 0 {
		return
	}

	// Step 1: Create leaf nodes
	leafNodes := []*Node{}
	for i := 0; i < len(sortedRecords); i += int(tree.Order) - 1 {
		end := i + int(tree.Order) - 1
		if end > len(sortedRecords) {
			end = len(sortedRecords)
		}

		leaf := tree.createLeafNode()
		for j := i; j < end; j++ {
			leaf.Keys = append(leaf.Keys, sortedRecords[j].Key)
			leaf.records = append(leaf.records, sortedRecords[j])
		}
		leaf.NumKeys = uint8(len(leaf.Keys))

		// Link leaf nodes
		if len(leafNodes) > 0 {
			leafNodes[len(leafNodes)-1].next = leaf
			leaf.prev = leafNodes[len(leafNodes)-1]
		}

		leafNodes = append(leafNodes, leaf)
	}

	// Step 2: Build internal nodes
	tree.root = buildInternalNodes(leafNodes, tree.Order)
}

// Recursively build internal nodes
func buildInternalNodes(children []*Node, order uint8) *Node {
	if len(children) == 1 {
		return children[0] // Root node
	}

	internalNodes := []*Node{}
	keys := [][]byte{}

	for i := 0; i < len(children); i += int(order) {
		end := i + int(order)
		if end > len(children) {
			end = len(children)
		}

		node := &Node{children: children[i:end]}
		for _, child := range children[i:end] {
			child.parent = node
		}

		// Pick first key from each child
		if i > 0 {
			keys = append(keys, children[i].Keys[0])
		}

		node.Keys = keys
		node.NumKeys = uint8(len(keys))
		internalNodes = append(internalNodes, node)
	}

	return buildInternalNodes(internalNodes, order)
}

//------------------------------------------------------------------

// Search searches for a key in the B+ tree using binary search.
func (tree *BTree) Search(key []byte) (*Record, error) {
	if tree == nil || tree.root == nil {
		return nil, ErrorTreeNil
	}

	node := tree.root

	// Traverse down to the leaf node
	for len(node.children) > 0 {
		index := binarySearch(node.Keys, key)
		node = node.children[index]
	}

	// Perform binary search in the leaf node
	index := binarySearch(node.Keys, key)
	if index < len(node.Keys) && bytes.Equal(node.Keys[index], key) {
		return node.records[index], nil
	}

	return nil, ErrorKeyNotFound
}

// Binary search helper function
func binarySearch(keys [][]byte, key []byte) int {
	return sort.Search(len(keys), func(i int) bool {
		return bytes.Compare(keys[i], key) >= 0
	})
}

// RangeScan retrieves all records in the range [startKey, endKey].
func (tree *BTree) RangeScan(startKey, endKey []byte) []*Record {
	if tree == nil || tree.root == nil {
		return nil
	}

	var results []*Record
	node := tree.root

	// Traverse down to the correct leaf node
	for len(node.children) > 0 {
		index := binarySearch(node.Keys, startKey)
		node = node.children[index]
	}

	// Scan through leaf nodes until we reach endKey
	for node != nil {
		for i := 0; i < len(node.Keys); i++ {
			if bytes.Compare(node.Keys[i], startKey) >= 0 && bytes.Compare(node.Keys[i], endKey) <= 0 {
				results = append(results, node.records[i])
			}
			// Stop if we exceed endKey
			if bytes.Compare(node.Keys[i], endKey) > 0 {
				return results
			}
		}
		node = node.next // Move to next leaf node
	}

	return results
}

//------------------------------------------------------------------

// deleteKey deletes a key from the B+ Tree.
func (tree *BTree) deleteKey(key []byte) {
	if tree.root == nil {
		fmt.Println("Tree is empty!")
		return
	}

	leaf := tree.findLeafNode(key)
	index := -1

	// Find the key in the leaf node
	for i, k := range leaf.Keys {
		if bytes.Equal(k, key) {
			index = i
			break
		}
	}
	if index == -1 {
		fmt.Println("Key not found!")
		return
	}

	// Remove the key and corresponding record
	leaf.Keys = append(leaf.Keys[:index], leaf.Keys[index+1:]...)
	leaf.records = append(leaf.records[:index], leaf.records[index+1:]...)
	leaf.NumKeys--

	// Handle underflow
	tree.handleUnderflow(leaf)
}

// handleUnderflow handles cases when a node has fewer than the required keys.
func (tree *BTree) handleUnderflow(node *Node) {
	minKeys := (int(tree.Order) - 1) / 2 // Minimum required keys
	if len(node.Keys) >= minKeys {
		return // No underflow
	}

	// Check if the node is the root
	if node == tree.root {
		if len(node.children) == 1 { // If root has only one child, make it the new root
			tree.root = node.children[0]
			tree.root.parent = nil
		}
		return
	}

	// Find the node's parent and its position
	parent := node.parent
	pos := 0
	for pos < len(parent.children) && parent.children[pos] != node {
		pos++
	}

	// Try to borrow from left sibling
	if pos > 0 {
		leftSibling := parent.children[pos-1]
		if len(leftSibling.Keys) > minKeys {
			// Borrow key from left sibling
			borrowedKey := leftSibling.Keys[len(leftSibling.Keys)-1]
			leftSibling.Keys = leftSibling.Keys[:len(leftSibling.Keys)-1]
			node.Keys = append([][]byte{borrowedKey}, node.Keys...)
			parent.Keys[pos-1] = borrowedKey
			return
		}
	}

	// Try to borrow from right sibling
	if pos < len(parent.children)-1 {
		rightSibling := parent.children[pos+1]
		if len(rightSibling.Keys) > minKeys {
			// Borrow key from right sibling
			borrowedKey := rightSibling.Keys[0]
			rightSibling.Keys = rightSibling.Keys[1:]
			node.Keys = append(node.Keys, borrowedKey)
			parent.Keys[pos] = rightSibling.Keys[0]
			return
		}
	}

	// Merge with left sibling
	if pos > 0 {
		leftSibling := parent.children[pos-1]
		leftSibling.Keys = append(leftSibling.Keys, node.Keys...)
		leftSibling.records = append(leftSibling.records, node.records...)
		parent.children = append(parent.children[:pos], parent.children[pos+1:]...)
		parent.Keys = append(parent.Keys[:pos-1], parent.Keys[pos:]...)
		tree.handleUnderflow(parent)
	} else { // Merge with right sibling
		rightSibling := parent.children[pos+1]
		node.Keys = append(node.Keys, rightSibling.Keys...)
		node.records = append(node.records, rightSibling.records...)
		parent.children = append(parent.children[:pos+1], parent.children[pos+2:]...)
		parent.Keys = append(parent.Keys[:pos], parent.Keys[pos+1:]...)
		tree.handleUnderflow(parent)
	}
}

// ------------------------------------------------------------------

// NodeJSON represents a node in a JSON-friendly structure
type NodeJSON struct {
	Key      []string   `json:"key"`
	Value    []string   `json:"value"`
	Children []NodeJSON `json:"children"`
}

// ConvertNodeToJSON recursively converts a Node into a JSON-friendly structure
func (node *Node) ConvertNodeToJSON() NodeJSON {
	if node == nil {
		return NodeJSON{}
	}

	keys := make([]string, len(node.Keys))
	values := make([]string, len(node.records))

	for i, key := range node.Keys {
		keys[i] = string(key)
	}
	for i, record := range node.records {
		values[i] = string(record.Value)
	}

	children := make([]NodeJSON, len(node.children))
	for i, child := range node.children {
		children[i] = child.ConvertNodeToJSON()
	}

	return NodeJSON{
		Key:      keys,
		Value:    values,
		Children: children,
	}
}

// ConvertBTreeToJSON converts the entire BTree into a JSON structure
func (tree *BTree) ConvertBTreeToJSON() ([]byte, error) {
	if tree.root == nil {
		return nil, nil
	}

	treeJSON := map[string]NodeJSON{
		"root": tree.root.ConvertNodeToJSON(),
	}

	jsonData, err := json.MarshalIndent(treeJSON, "", "  ")
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
