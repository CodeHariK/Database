package secretary

import (
	"bytes"
	"sort"

	"github.com/codeharik/secretary/utils"
)

//------------------------------------------------------------------
// Node Verification
//------------------------------------------------------------------

/*
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
func (tree *BTree) NodeVerify(node *Node) error {
	// A node is either a leaf (has records) or an internal node (has children), not both
	if (node.records != nil && node.children != nil) || (node.records == nil && node.children == nil) {
		return ErrorNodeIsEitherLeaforInternal
	}

	// Check len(keys)
	if len(node.Keys) >= int(tree.Order) {
		return ErrorKeysGTEOrder
	}
	if node != tree.root && len(node.Keys) < int(tree.minNumKeys) {
		return ErrorKeysLTOrder
	}

	// Check if Parent knows Node
	if node.parent != nil {
		contains := false
		for _, c := range node.parent.children {
			if c == node {
				contains = true
			}
		}
		if !contains {
			return ErrorParentNotKnowChild(node)
		}
	}

	// Check Next and Prev Pointers
	if node.next != nil && node.next.prev != node {
		return ErrorNextNodeLink(node)
	}
	if node.prev != nil && node.prev.next != node {
		return ErrorPrevNodeLink(node)
	}

	// Internal node, check len(children), check Children knows Node, Check MinKeys
	if node.children != nil {
		if len(node.children) != (len(node.Keys) + 1) {
			return ErrorInternalLenChildren
		}

		for i, child := range node.children {
			if child.parent != node {
				return ErrorChildNotKnowParent(node, child)
			}

			minLeafKey, err := child.getMinLeafKey()

			if i > 0 && err == nil && bytes.Compare(minLeafKey, node.Keys[i-1]) != 0 {
				return ErrorNodeMinKeyMismatch
			}
		}
	}

	// Leaf node, check len(records), check nodekey match recordkey,
	if node.records != nil {
		if len(node.records) != len(node.Keys) {
			return ErrorLeafLenRecords
		}
		for i, r := range node.records {
			if bytes.Compare(r.Key, node.Keys[i]) != 0 {
				return ErrorRecordKeyMismatch
			}
		}
		if !areRecordsSorted(node.records) {
			return ErrorRecordsNotSorted
		}
	}

	// Validate each key size and key offset
	for i, el := range node.Keys {
		if len(el) != KEY_SIZE {
			return ErrorInvalidKey
		}

		// Are keys sorted
		if i >= 1 && bytes.Compare(node.Keys[i-1], node.Keys[i]) >= 0 {
			return ErrorKeysNotOrdered
		}

		// // Bounds check for KeyOffsets
		// if i < len(n.KeyOffsets) {
		// 	if err := tree.dataLocationCheck(n.KeyOffsets[i]); err != nil {
		// 		logFatalWithStack(n.NodeID, err)
		// 	}
		// }
	}

	return nil
}

func (tree *BTree) recursiveNodeVerify(node *Node) []error {
	rErrs := []error{}
	if node != nil {
		err := tree.NodeVerify(node)
		if err != nil {
			rErrs = append(rErrs, err)
		}
		for _, n := range node.children {
			cErrs := tree.recursiveNodeVerify(n)
			if cErrs != nil {
				rErrs = append(rErrs, cErrs...)
			}
		}
	}
	if len(rErrs) == 0 {
		return nil
	}
	return rErrs
}

func (tree *BTree) TreeVerify() []error {
	return tree.recursiveNodeVerify(tree.root)
}

//------------------------------------------------------------------
// Create Nodes
//------------------------------------------------------------------

// Create a new internal node
func (tree *BTree) createInternalNode(children []*Node) *Node {
	tree.NodeSeq += 1

	if children == nil {
		children = make([]*Node, 0)
	}

	return &Node{
		Keys:     make([][]byte, 0),
		children: children,
		// NumKeys:  0,

		NodeID: tree.NodeSeq,
	}
}

// Create a new leaf node
func (tree *BTree) createLeafNode() *Node {
	tree.NodeSeq += 1

	return &Node{
		Keys:    make([][]byte, 0),
		records: make([]*Record, 0),
		// NumKeys: 0,

		NodeID: tree.NodeSeq,
	}
}

//------------------------------------------------------------------
// Set/Update Key
//------------------------------------------------------------------

// Set key-value in leaf node
func (leaf *Node) setLeafKV(key []byte, value []byte) {
	i, _ := leaf.getKey(key)

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

	tree.setParentKV(leaf, newLeaf.Keys[0], newLeaf)
}

func (tree *BTree) setParentKV(left *Node, key []byte, right *Node) {
	// If left is leaf and root, then create new root
	if left.parent == nil {
		newRoot := tree.createInternalNode(nil)
		newRoot.Keys = [][]byte{key}
		newRoot.children = []*Node{left, right}
		// newRoot.NumKeys = uint8(len(newRoot.Keys))

		left.parent = newRoot
		right.parent = newRoot
		tree.root = newRoot

		return
	}

	parent := left.parent
	setIdx, _ := parent.getKey(key)

	parent.Keys = append(
		parent.Keys[:setIdx],
		append(
			[][]byte{key},
			parent.Keys[setIdx:]...,
		)...)

	parent.children = append(
		parent.children[:setIdx+1],
		append(
			[]*Node{right},
			parent.children[setIdx+1:]...,
		)...)

	right.parent = parent
	// parent.NumKeys++

	// if parent.NumKeys >= tree.Order {
	if len(parent.Keys) >= int(tree.Order) {
		tree.splitInternal(parent)
	}
}

// Split an internal node
func (tree *BTree) splitInternal(node *Node) {
	mid := len(node.Keys) / 2

	newRightInternal := tree.createInternalNode(nil)
	newRightInternal.Keys = append(
		newRightInternal.Keys,
		node.Keys[mid+1:]...)
	newRightInternal.children = append(
		newRightInternal.children,
		node.children[mid+1:]...)
	// newInternal.NumKeys = uint8(len(newInternal.Keys))

	for _, child := range newRightInternal.children {
		child.parent = newRightInternal
	}

	promotedKey := node.Keys[mid]
	node.Keys = node.Keys[:mid]
	node.children = node.children[:mid+1]
	// node.NumKeys = uint8(len(node.Keys))

	newRightInternal.next = node.next
	node.next = newRightInternal
	newRightInternal.prev = node

	tree.setParentKV(node, promotedKey, newRightInternal)
}

// Set a key-value pair into the B+ Tree
func (tree *BTree) Set(key []byte, value []byte) error {
	if len(key) != 16 {
		return ErrorInvalidKey
	}

	if tree.root == nil {
		tree.root = tree.createLeafNode()
		tree.root.setLeafKV(key, value)

		return nil
	}

	leaf, index, found := tree.getLeafNode(key)
	if found && bytes.Compare(leaf.Keys[index], key) == 0 {
		return ErrorDuplicateKey
	}

	leaf.setLeafKV(key, value)

	if len(leaf.Keys) >= int(tree.Order) {
		tree.splitLeaf(leaf)
	}

	return nil
}

// Update a key-value pair in the B+ Tree
func (tree *BTree) Update(key []byte, value []byte) error {
	if len(key) != 16 {
		return ErrorInvalidKey
	}

	leaf, index, found := tree.getLeafNode(key)
	if found {
		leaf.records[index].Value = value
		return nil
	}
	return ErrorKeyNotFound
}

//------------------------------------------------------------------
// Sorted Records Set
//------------------------------------------------------------------

func areRecordsSorted(records []*Record) bool {
	for i := 1; i < len(records); i++ {
		if bytes.Compare(records[i-1].Key, records[i].Key) > 0 {
			return false // Not sorted
		}
	}
	return true // Sorted
}

// SortedRecordSet: set sorted records into the B+ Tree efficiently
func (tree *BTree) SortedRecordSet(sortedRecords []*Record) error {
	if !areRecordsSorted(sortedRecords) || len(sortedRecords) == 0 {
		return ErrorRecordsNotSorted
	}

	leafNodes := tree.buildSortedLeafNodes(sortedRecords)

	tree.root = tree.buildInternalNodes(leafNodes)

	return nil
}

func (tree *BTree) buildSortedLeafNodes(sortedRecords []*Record) []*Node {
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

		leafNodes = append(leafNodes, leaf)
	}

	return leafNodes
}

// Recursively build internal nodes
func (tree *BTree) buildInternalNodes(children []*Node) *Node {
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
		if (len(children)-start)%int(tree.Order) == 1 {
			end = start + int(tree.Order) - 1
		} else {
			end = start + int(tree.Order)
		}

		if end > len(children) {
			end = len(children)
		}

		// Create an internal node from this chunk of children
		node := tree.createInternalNode(children[start:end])

		// Assign separator keys (first key of each child, except the first one)
		// Link parent-child relationships
		for i, child := range children[start:end] {
			child.parent = node

			minLeafKey, err := child.getMinLeafKey()
			if err == nil && i != 0 {
				node.Keys = append(node.Keys, minLeafKey)
			}
		}

		internalNodes = append(internalNodes, node)

		if (len(children)-start)%int(tree.Order) == 1 {
			start += int(tree.Order) - 1
		} else {
			start += int(tree.Order)
		}
	}

	return tree.buildInternalNodes(internalNodes)
}

//------------------------------------------------------------------
// Get Key
//------------------------------------------------------------------

// Binary search helper function
func (n *Node) getKey(key []byte) (index int, found bool) {
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

func (node *Node) getMinLeafKey() ([]byte, error) {
	if len(node.children) == 0 {
		if len(node.Keys) > 0 {
			return node.Keys[0], nil
		}
		return nil, ErrorKeyNotFound
	}
	return node.children[0].getMinLeafKey()
}

// Find the appropriate leaf node
func (tree *BTree) getLeafNode(key []byte) (node *Node, index int, found bool) {
	node = tree.root

	// Traverse internal nodes
	for len(node.children) > 0 {
		index, found := node.getKey(key)
		if found {
			node = node.children[index+1]
		} else {
			node = node.children[index]
		}
	}

	// Search within the leaf node
	index, found = node.getKey(key)

	return node, index, found
}

func (tree *BTree) Get(key []byte) (*Record, error) {
	node, index, found := tree.getLeafNode(key)
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

	startNode, startIndex, _ := tree.getLeafNode(startKey)
	endNode, endIndex, endFound := tree.getLeafNode(endKey)

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
// Delete
//------------------------------------------------------------------

// Delete deletes a key from the B+ Tree.
func (tree *BTree) Delete(key []byte) error {
	if tree == nil || tree.root == nil {
		return ErrorTreeNil
	}

	leaf, index, found := tree.getLeafNode(key)

	if !found {
		return ErrorKeyNotFound
	}

	// // If first key of leaf is deleted && leaf is not first child of its parent, then update parent key
	// if index == 0 && len(leaf.Keys) > 1 && bytes.Compare(key, leaf.parent.Keys[0]) >= 0 {
	// 	pi, pf := leaf.parent.getKey(key)
	// 	if !pf {
	// 		return errors.New("Leaf Key 0 not in Parent Keys")
	// 	}
	// 	fmt.Println("Removed child parent seperator key")
	// 	leaf.parent.Keys[pi] = leaf.Keys[index+1]
	// }

	// Remove the key and corresponding record
	leaf.Keys = append(leaf.Keys[:index], leaf.Keys[index+1:]...)
	leaf.records = append(leaf.records[:index], leaf.records[index+1:]...)
	// leaf.NumKeys--

	utils.Log("key", string(key), "leaf", leaf.NodeID, "index", index, "found", found)

	// Handle underflow
	tree.handleUnderflow(leaf)

	tree.recursiveFixInternalNodeChildLinks(leaf)

	return nil
}

// handleUnderflow handles cases when a node has fewer than the required keys.
func (tree *BTree) handleUnderflow(node *Node) {
	minKeys := (int(tree.Order) - 1) / 2 // Minimum required keys
	if len(node.Keys) >= minKeys {
		return // No underflow
	}

	utils.Log("handleUnderflow", node.NodeID, "node.keys", utils.ArrayToStrings(node.Keys))

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
		// leftSibling := node.prev
		leftSibling := parent.children[pos-1]

		utils.Log(
			"**** Try to borrow from leftSibling", leftSibling.NodeID,
			"len(leftSibling.Keys)", len(leftSibling.Keys),
			"minKeys", minKeys,
			"len(leftSibling.Keys) > minKeys)", len(leftSibling.Keys) > minKeys,
		)

		if len(leftSibling.Keys) > minKeys {
			// Borrow key from left sibling
			blen := len(leftSibling.Keys) - 1
			borrowedKey := leftSibling.Keys[blen]
			leftSibling.Keys = leftSibling.Keys[:blen]

			node.Keys = append([][]byte{borrowedKey}, node.Keys...)

			parent.Keys[pos-1] = borrowedKey

			if leftSibling.children == nil {
				rlen := len(leftSibling.records) - 1
				borrowedRecord := leftSibling.records[rlen]

				utils.Log(
					"BorrowedRecord", rlen,
					"string(borrowedKey)", string(borrowedKey),
					"string(borrowedRecord.Value)", string(borrowedRecord.Value),
				)

				leftSibling.records = leftSibling.records[:rlen]
				node.records = append([]*Record{borrowedRecord}, node.records...)
			} else {
				clen := len(leftSibling.children) - 1
				borrowedChild := leftSibling.children[clen]

				utils.Log("BorrowedChild", borrowedChild.NodeID, "borrowedKey", string(borrowedKey))

				leftSibling.children = leftSibling.children[:clen]
				node.children = append([]*Node{borrowedChild}, node.children...)

				borrowedChild.parent = node

				// recursiveFixInternalNodeChildLinks(node)
			}

			// recursiveFixInternalNodeChildLinks(parent)
			tree.recursiveFixInternalNodeChildLinks(node)

			utils.Log("**** Borrow from leftSibling ", tree.NodeToJSON(leftSibling, 2),
				"BorrowedKey:", string(borrowedKey),
				"parent", tree.NodeToJSON(parent, 2))

			return
		}
	}

	// Try to borrow from right sibling
	if pos < len(parent.children)-1 {
		rightSibling := parent.children[pos+1]
		// rightSibling := node.next

		utils.Log(
			"**** Try to borrow from rightSibling", rightSibling.NodeID,
			"len(rightSibling.Keys)", len(rightSibling.Keys),
			"minKeys", minKeys,
			"len(rightSibling.Keys) > minKeys", len(rightSibling.Keys) > minKeys,
		)

		if len(rightSibling.Keys) > minKeys {
			// Borrow key from right sibling

			borrowedKey := rightSibling.Keys[0]
			rightSibling.Keys = rightSibling.Keys[1:]

			node.Keys = append(node.Keys, borrowedKey)

			// parent.Keys[pos] = rightSibling.Keys[0]
			// parent.Keys[pos-1] = node.Keys[0]

			if rightSibling.children == nil {
				borrowedRecord := rightSibling.records[0]

				utils.Log("BorrowedRecord", borrowedRecord)

				rightSibling.records = rightSibling.records[1:]
				node.records = append(node.records, borrowedRecord)
			} else {
				borrowedChild := rightSibling.children[0]

				utils.Log("BorrowedChild", borrowedChild)

				rightSibling.children = rightSibling.children[1:]
				node.children = append(node.children, borrowedChild)

				borrowedChild.parent = node

				// recursiveFixInternalNodeChildLinks(node)
			}

			// recursiveFixInternalNodeChildLinks(parent)
			tree.recursiveFixInternalNodeChildLinks(node)

			utils.Log("**** Borrow from rightSibling ", tree.NodeToJSON(rightSibling, 2),
				"BorrowedKey:", string(borrowedKey),
				"parent", tree.NodeToJSON(parent, 2))

			return
		}
	}

	// Merge with left sibling
	if pos > 0 {
		leftSibling := parent.children[pos-1]
		// leftSibling := node.prev

		leftSibling.Keys = append(leftSibling.Keys, node.Keys...)

		if leftSibling.children == nil {
			leftSibling.records = append(leftSibling.records, node.records...)
		} else {
			leftSibling.children = append(leftSibling.children, node.children...)
		}

		// parent.Keys = append(parent.Keys[:pos-1], parent.Keys[pos:]...)
		parent.children = append(parent.children[:pos], parent.children[pos+1:]...)

		// for _, c := range leftSibling.children {
		// 	c.parent = leftSibling
		// 	utils.Log("leftSibling", leftSibling.NodeID, "childID", c.NodeID, "child.parent", c.parent.NodeID)
		// }

		node.parent = nil

		if node.prev != nil {
			leftSibling.next = node.next
		}
		if node.next != nil {
			node.next.prev = leftSibling
		}

		utils.Log("**** Merge with left sibling -> Pos", pos,
			"Parent", tree.NodeToJSON(parent, 2),
			"Node", tree.NodeToJSON(node, 2),
			"leftSibling", tree.NodeToJSON(leftSibling, 2),
		)

		// recursiveFixInternalNodeChildLinks(parent)
		tree.recursiveFixInternalNodeChildLinks(leftSibling)

		tree.handleUnderflow(parent)
	} else
	// Merge right sibling
	{
		rightSibling := parent.children[pos+1]
		// rightSibling := node.next

		node.Keys = append(node.Keys, rightSibling.Keys...)

		if rightSibling.children == nil {
			node.records = append(node.records, rightSibling.records...)
		} else {
			node.children = append(node.children, rightSibling.children...)
		}

		// parent.Keys = append(parent.Keys[:pos], parent.Keys[pos+1:]...)
		parent.children = append(parent.children[:pos+1], parent.children[pos+2:]...)

		if rightSibling.prev != nil {
			rightSibling.prev.next = rightSibling.next
		}
		if rightSibling.next != nil {
			rightSibling.next.prev = rightSibling.prev
		}

		utils.Log("**** Merge right sibling -> Pos", pos,
			"Parent", tree.NodeToJSON(parent, 2),
			"Node", tree.NodeToJSON(node, 2),
			"rightSibling", tree.NodeToJSON(rightSibling, 2),
		)

		// recursiveFixInternalNodeChildLinks(parent)
		tree.recursiveFixInternalNodeChildLinks(node)

		tree.handleUnderflow(parent)
	}
}

func (tree *BTree) fixInternalNodeChildLinks(node *Node) {
	if node != nil && node.children != nil {
		node.Keys = [][]byte{}
		for i, child := range node.children {
			minLeafKey, err := child.getMinLeafKey()
			if i > 0 && err == nil {
				node.Keys = append(node.Keys, minLeafKey)
			}
			child.parent = node
		}
	}
}

func (tree *BTree) recursiveFixInternalNodeChildLinks(node *Node) {
	tree.fixInternalNodeChildLinks(node)
	if node.parent != nil {
		tree.recursiveFixInternalNodeChildLinks(node.parent)
	}
}

//------------------------------------------------------------------
// Node/Tree to JSON
//------------------------------------------------------------------

// NodeJSON represents a node in a JSON-friendly structure
type NodeJSON struct {
	NodeId   uint64 `json:"nodeID"`
	NextId   uint64 `json:"nextID"`
	PrevId   uint64 `json:"prevID"`
	ParentId uint64 `json:"parentID"`

	// NumKeys  uint8      `json:"numKeys"`

	Key      []string   `json:"key"`
	Value    []string   `json:"value"`
	Children []NodeJSON `json:"children"`

	Errors []string `json:"errors"`
}

// NodeToJSON recursively converts a Node into a JSON-friendly structure
func (tree *BTree) NodeToJSON(node *Node, height int) NodeJSON {
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

	var children []NodeJSON
	if height != 1 {
		children = make([]NodeJSON, len(node.children))
		for i, child := range node.children {
			children[i] = tree.NodeToJSON(child, height-1)
		}
	}

	nextId := uint64(0)
	prevId := uint64(0)
	parentId := uint64(0)
	if node.next != nil {
		nextId = node.next.NodeID
	}
	if node.prev != nil {
		prevId = node.prev.NodeID
	}
	if node.parent != nil {
		parentId = node.parent.NodeID
	}

	return NodeJSON{
		NodeId:   node.NodeID,
		NextId:   nextId,
		PrevId:   prevId,
		ParentId: parentId,
		Key:      keys,
		Value:    values,
		Children: children,
		// NumKeys:  node.NumKeys,

		Errors: utils.ArrayToStrings(tree.recursiveNodeVerify(node)),
	}
}

func (tree *BTree) ToJSON() NodeJSON {
	return tree.NodeToJSON(tree.root, tree.Height())
}
