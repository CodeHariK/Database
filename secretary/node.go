package secretary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"sort"

	"github.com/codeharik/secretary/utils/binstruct"
)

func logFatalWithStack(msg ...any) {
	log.Println(string(debug.Stack()))
	log.Fatal(msg...)
}

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
func (tree *BTree) NodeVerify(n *Node) {
	// A node is either a leaf (has records) or an internal node (has children), not both
	if (n.records != nil && n.children != nil) || (n.records == nil && n.children == nil) {
		logFatalWithStack(n.NodeID, ErrorNodeIsEitherLeaforInternal)
	}

	if len(n.Keys) >= int(tree.Order) {
		// if len(n.Keys) != int(n.NumKeys) || n.NumKeys >= tree.Order {
		logFatalWithStack(n.NodeID, ErrorNumKeysNotMatching, len(n.Keys), tree.Order)
	}

	if n.parent != nil {
		contains := false
		for _, c := range n.parent.children {
			if c == n {
				contains = true
			}
		}
		if !contains {
			logFatalWithStack("child:", n.NodeID, " parent:", n.parent.NodeID, ErrorInvalidNodeLink)
		}
	}

	if (n.next != nil && n.next.prev != n) || (n.prev != nil && n.prev.next != n) {
		logFatalWithStack("node:", n.NodeID, "nextprev:", n.next.prev.NodeID, " prevnext:", n.prev.next.NodeID, ErrorInvalidNodeLink)
	}

	if n.children != nil {
		// Internal node
		if len(n.children) != (len(n.Keys) + 1) {
			// if len(n.children) != int(n.NumKeys+1) {

			for _, c := range n.children {
				if c.parent != n {
					logFatalWithStack("child:", c.NodeID, " parent:", n.NodeID, ErrorInvalidNodeLink, len(n.Keys), tree.Order, len(n.children))
				}
			}

			logFatalWithStack("child ", n.NodeID, ErrorNumKeysNotMatching, len(n.Keys), tree.Order, len(n.children))
		}
	} else if n.records != nil {
		// Leaf node
		if len(n.records) != len(n.Keys) {

			for i, r := range n.records {
				if bytes.Compare(r.Key, n.Keys[i]) != 0 {
					logFatalWithStack("record ", r.Key, ErrorInvalidKey, len(n.Keys), tree.Order, len(n.records))
				}
			}

			logFatalWithStack("leaf ", n.NodeID, ErrorNumKeysNotMatching, len(n.Keys), tree.Order, len(n.records))
		}
	}

	// Validate each key size and key offset
	for i, el := range n.Keys {
		if len(el) != KEY_SIZE {
			logFatalWithStack(n.NodeID, ErrorInvalidKey)
		}

		if i >= 1 && bytes.Compare(n.Keys[i-1], n.Keys[i]) >= 0 {
			logFatalWithStack("keys out of order in node ", n.NodeID)
		}

		// // Bounds check for KeyOffsets
		// if i < len(n.KeyOffsets) {
		// 	if err := tree.dataLocationCheck(n.KeyOffsets[i]); err != nil {
		// 		logFatalWithStack(n.NodeID, err)
		// 	}
		// }
	}
}

func (tree *BTree) RecursiveNodeVerify(node *Node) {
	if node != nil {
		tree.NodeVerify(node)

		for _, n := range node.children {
			tree.RecursiveNodeVerify(n)
		}
	}
}

func (tree *BTree) TreeVerify() {
	tree.RecursiveNodeVerify(tree.root)
}

// Create a new internal node
func (tree *BTree) createInternalNode(children []*Node) *Node {
	tree.NumNodes += 1

	if children == nil {
		children = make([]*Node, 0)
	}

	return &Node{
		Keys:     make([][]byte, 0),
		children: children,
		// NumKeys:  0,

		NodeID: tree.NumNodes,
	}
}

// Create a new leaf node
func (tree *BTree) createLeafNode() *Node {
	tree.NumNodes += 1

	return &Node{
		Keys:    make([][]byte, 0),
		records: make([]*Record, 0),
		// NumKeys: 0,

		NodeID: tree.NumNodes,
	}
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

//------------------------------------------------------------------

// Insert key-value into a leaf node
func (tree *BTree) insertIntoLeaf(leaf *Node, key []byte, value []byte) {
	i, _ := leaf.SearchKey(key)

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
	// leaf.NumKeys++
}

// Split a leaf node
func (tree *BTree) splitLeaf(leaf *Node) {
	mid := len(leaf.Keys) / 2
	newLeaf := tree.createLeafNode()

	newLeaf.Keys = append(newLeaf.Keys, leaf.Keys[mid:]...)
	newLeaf.records = append(newLeaf.records, leaf.records[mid:]...)
	// newLeaf.NumKeys = uint8(len(newLeaf.Keys))

	leaf.Keys = leaf.Keys[:mid]
	leaf.records = leaf.records[:mid]
	// leaf.NumKeys = uint8(len(leaf.Keys))

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

	tree.NodeVerify(leaf)
	tree.NodeVerify(newLeaf)
}

// Insert into an internal node
func (tree *BTree) insertIntoParent(left *Node, key []byte, right *Node) {
	// If left is leaf and root, then create new root
	if left.parent == nil {
		newRoot := tree.createInternalNode(nil)
		newRoot.Keys = [][]byte{key}
		newRoot.children = []*Node{left, right}
		// newRoot.NumKeys = uint8(len(newRoot.Keys))

		left.parent = newRoot
		right.parent = newRoot
		tree.root = newRoot

		tree.NodeVerify(left)
		tree.NodeVerify(right)
		tree.NodeVerify(tree.root)
		return
	}

	parent := left.parent
	insertIdx, _ := parent.SearchKey(key)

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
	// parent.NumKeys++

	// fmt.Println("intl ", parent.NodeID, len(parent.Keys), "split:", len(parent.Keys) >= int(tree.Order))

	// if parent.NumKeys >= tree.Order {
	if len(parent.Keys) >= int(tree.Order) {
		tree.splitInternal(parent)
	}

	tree.NodeVerify(parent)
}

// Split an internal node
func (tree *BTree) splitInternal(node *Node) {
	mid := len(node.Keys) / 2

	newInternal := tree.createInternalNode(nil)
	newInternal.Keys = append(
		newInternal.Keys,
		node.Keys[mid+1:]...)
	newInternal.children = append(
		newInternal.children,
		node.children[mid+1:]...)
	// newInternal.NumKeys = uint8(len(newInternal.Keys))

	for _, child := range newInternal.children {
		child.parent = newInternal
	}

	promotedKey := node.Keys[mid]
	node.Keys = node.Keys[:mid]
	node.children = node.children[:mid+1]
	// node.NumKeys = uint8(len(node.Keys))

	tree.insertIntoParent(node, promotedKey, newInternal)

	tree.NodeVerify(node)
	tree.NodeVerify(newInternal)
}

// Insert a key-value pair into the B+ Tree
func (tree *BTree) Insert(key []byte, value []byte) error {
	if len(key) != 16 {
		return ErrorInvalidKey
	}

	if tree.root == nil {
		tree.root = tree.createLeafNode()
		tree.insertIntoLeaf(tree.root, key, value)

		tree.NodeVerify(tree.root)
		return nil
	}

	leaf, index, found := tree.SearchLeafNode(key)
	if found && bytes.Compare(leaf.Keys[index], key) == 0 {
		return ErrorDuplicateKey
	}

	tree.insertIntoLeaf(leaf, key, value)

	// fmt.Println("leaf ", leaf.NodeID, len(leaf.Keys), "split:", len(leaf.Keys) >= int(tree.Order))

	if len(leaf.Keys) >= int(tree.Order) {
		tree.splitLeaf(leaf)
	}

	tree.NodeVerify(leaf)
	return nil
}

// Update a key-value pair in the B+ Tree
func (tree *BTree) Update(key []byte, value []byte) error {
	if len(key) != 16 {
		return ErrorInvalidKey
	}

	leaf, index, found := tree.SearchLeafNode(key)
	if found {
		leaf.records[index].Value = value
		return nil
	}
	return ErrorKeyNotFound
}

// ------------------------------------------------------------------

// SortedRecordLoad inserts sorted records into the B+ Tree efficiently
func (tree *BTree) SortedRecordLoad(sortedRecords []*Record) {
	leafNodes := tree.buildSortedLeafNodes(sortedRecords)

	tree.root = tree.buildInternalNodes(leafNodes)

	tree.NodeVerify(tree.root)
}

func (tree *BTree) buildSortedLeafNodes(sortedRecords []*Record) []*Node {
	fmt.Println(len(sortedRecords))
	for _, r := range sortedRecords {
		fmt.Println(string(r.Key), string(r.Value))
	}
	fmt.Println("---------")

	if len(sortedRecords) == 0 {
		return []*Node{}
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
		// leaf.NumKeys = uint8(len(leaf.Keys))

		// Link leaf nodes
		if len(leafNodes) > 0 {
			leafNodes[len(leafNodes)-1].next = leaf
			leaf.prev = leafNodes[len(leafNodes)-1]
		}

		tree.NodeVerify(leaf)

		leafNodes = append(leafNodes, leaf)
	}

	fmt.Println(len(leafNodes))
	for _, n := range leafNodes {
		fmt.Println(n.ConvertNodeToJSON())
	}
	fmt.Println("---------")
	return leafNodes
}

// Recursively build internal nodes
func (tree *BTree) buildInternalNodes(children []*Node) *Node {
	fmt.Println(len(children))
	if len(children) == 1 {
		return children[0] // Root node
	}

	internalNodes := []*Node{}

	// group cant have less than 2 and more than tree.Order, there has to be 1 key atleast

	// Order 4 Keys : 3  Children : 1 : [1]
	// Order 4 Keys : 5  Children : 2 : [2]
	// Order 4 Keys : 8  Children : 3 : [3]
	// Order 4 Keys : 12 Children : 4 : [4]
	// Order 4 Keys : 14 Children : 5 : [3,2]
	// Order 4 Keys : 16 Children : 6 : [4,2]
	// Order 4 Keys : 20 Children : 7 : [4,3]
	// Order 4 Keys : 24 Children : 8 : [4,4]
	// Order 4 Keys : 26 Children : 9 : [4,3,2]

	for start := 0; start < len(children); {
		var end int
		if (len(children)-start)/int(tree.Order) == 1 {
			end = start + int(tree.Order) - 1
		} else {
			end = start + int(tree.Order)
		}

		if end > len(children) {
			end = len(children)
		}

		fmt.Println("order:", tree.Order, "start:", start, " end:", end, " len:", len(children))

		// Create an internal node from this chunk of children
		node := tree.createInternalNode(children[start:end])

		// Assign separator keys (first key of each child, except the first one)
		// Link parent-child relationships
		for k, child := range children[start:end] {
			child.parent = node

			if k != 0 {
				node.Keys = append(node.Keys, child.Keys[0])
			}
		}

		// s, _ := json.MarshalIndent(node.ConvertNodeToJSON(), "", "  ")
		// fmt.Print(string(s), "\n\n\n*******\n\n\n")

		tree.NodeVerify(node)

		internalNodes = append(internalNodes, node)

		if (len(children)-start)/int(tree.Order) == 1 {
			start += int(tree.Order) - 1
		} else {
			start += int(tree.Order)
		}
	}

	return tree.buildInternalNodes(internalNodes)
}

//------------------------------------------------------------------

// Binary search helper function
func (n *Node) SearchKey(key []byte) (index int, found bool) {
	index = sort.Search(
		len(n.Keys),
		func(i int) bool {
			return bytes.Compare(n.Keys[i], key) >= 0
		},
	)

	// Check if the key exists at the found index
	found = index < len(n.Keys) && bytes.Equal(n.Keys[index], key)

	return index, found
}

// Find the appropriate leaf node
func (tree *BTree) SearchLeafNode(key []byte) (node *Node, index int, found bool) {
	node = tree.root

	// fmt.Println(node.NodeID, utils.BytesToStrings(node.Keys))

	// Traverse internal nodes
	for len(node.children) > 0 {
		index, found := node.SearchKey(key)
		if found {
			node = node.children[index+1]
		} else {
			node = node.children[index]
		}
		// fmt.Println(index, found, node.NodeID, utils.BytesToStrings(node.Keys))
	}

	// Search within the leaf node
	index, found = node.SearchKey(key)
	// fmt.Println(index, found, node.NodeID, utils.BytesToStrings(node.Keys))

	return node, index, found
}

func (tree *BTree) SearchRecord(key []byte) (*Record, error) {
	node, index, found := tree.SearchLeafNode(key)
	if found {
		return node.records[index], nil
	}
	return nil, ErrorKeyNotFound
}

// RangeScan retrieves all records in the range [startKey, endKey].
func (tree *BTree) RangeScan(startKey, endKey []byte) []*Record {
	if tree == nil || tree.root == nil {
		return nil
	}

	var results []*Record

	startNode, startIndex, _ := tree.SearchLeafNode(startKey)
	endNode, endIndex, endFound := tree.SearchLeafNode(endKey)

	// Iterate over nodes
	for node := startNode; node != nil; node = node.next {
		// Determine the range of indices to iterate over
		start := startIndex
		end := len(node.records)
		if node == endNode {
			end = endIndex // Include records up to endIndex
			if endFound {
				end = endIndex + 1
			}
		}

		// Iterate over records within the node
		for i := start; i < end; i++ {
			record := node.records[i]
			results = append(results, record)
		}

		// Reset startIndex for the next node
		startIndex = 0
		if node == endNode {
			break
		}
	}

	return results
}

//------------------------------------------------------------------

// // Delete deletes a key from the B+ Tree.
// func (tree *BTree) Delete(key []byte) error {
// 	if tree == nil || tree.root == nil {
// 		return ErrorTreeNil
// 	}

// 	leaf := tree.searchLeafNode(key)
// 	index := -1

// 	// Find the key in the leaf node
// 	for i, k := range leaf.Keys {
// 		if bytes.Equal(k, key) {
// 			index = i
// 			break
// 		}
// 	}
// 	if index == -1 {
// 		return ErrorKeyNotFound
// 	}

// 	// Remove the key and corresponding record
// 	leaf.Keys = append(leaf.Keys[:index], leaf.Keys[index+1:]...)
// 	leaf.records = append(leaf.records[:index], leaf.records[index+1:]...)
// 	// leaf.NumKeys--

// 	tree.NodeCheck(leaf)

// 	// Handle underflow
// 	tree.handleUnderflow(leaf)

// 	return nil
// }

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

		tree.NodeVerify(tree.root)
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

			tree.NodeVerify(parent)
			tree.NodeVerify(leftSibling)

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

			tree.NodeVerify(parent)
			tree.NodeVerify(rightSibling)

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

		tree.NodeVerify(parent)
		tree.NodeVerify(leftSibling)

	} else { // Merge with right sibling
		rightSibling := parent.children[pos+1]
		node.Keys = append(node.Keys, rightSibling.Keys...)
		node.records = append(node.records, rightSibling.records...)
		parent.children = append(parent.children[:pos+1], parent.children[pos+2:]...)
		parent.Keys = append(parent.Keys[:pos], parent.Keys[pos+1:]...)
		tree.handleUnderflow(parent)

		tree.NodeVerify(parent)
		tree.NodeVerify(rightSibling)
	}
}

// ------------------------------------------------------------------

// NodeJSON represents a node in a JSON-friendly structure
type NodeJSON struct {
	NodeId uint64 `json:"nodeID"`
	NextId uint64 `json:"nextID"`
	PrevId uint64 `json:"prevID"`
	// NumKeys  uint8      `json:"numKeys"`
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

	nextId := uint64(0)
	prevId := uint64(0)
	if node.next != nil {
		nextId = node.next.NodeID
	}
	if node.prev != nil {
		prevId = node.prev.NodeID
	}

	return NodeJSON{
		NodeId:   node.NodeID,
		NextId:   nextId,
		PrevId:   prevId,
		Key:      keys,
		Value:    values,
		Children: children,
		// NumKeys:  node.NumKeys,
	}
}

// ConvertBTreeToJSON converts the entire BTree into a JSON structure
func (tree *BTree) ConvertBTreeToJSON() ([]byte, error) {
	if tree.root == nil {
		return nil, ErrorTreeNil
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
