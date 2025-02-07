package secretary

import (
	"bytes"
	"sort"

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

// TODO : Binary search key
func (tree *BTree) searchNode(n *Node, key []byte) (*Node, error) {
	if len(key) != KEY_SIZE {
		return nil, ErrorInvalidKey
	}

	tree.searchKey(tree.root, key)

	return nil, ErrorNodeNotInTree
}

func (tree *BTree) dataLocationCheck(location DataLocation) error {
	if location == -1 {
		return ErrorInvalidDataLocation
	}
	return nil
}

func (tree *BTree) addKey(n *Node, key []byte, keyOffset DataLocation) error {
	if (n.NumKeys + 1) > tree.Order {
		return ErrorNumKeysMoreThanOrder
	}
	if len(key) != KEY_SIZE {
		return ErrorInvalidKey
	}
	if err := tree.dataLocationCheck(keyOffset); err != nil {
		return err
	}

	n.NumKeys += 1

	n.KeyOffsets = append(n.KeyOffsets, keyOffset)
	n.Keys = append(n.Keys, key)

	return nil
}

// TODO : Binary search key
// Returns equal or less than
func (tree *BTree) searchKey(n *Node, key []byte) (int, error) {
	if len(key) != KEY_SIZE {
		return -1, ErrorInvalidKey
	}

	for i, k := range n.Keys {
		if bytes.Compare(key, k) == 0 {
			return i, nil
		}
	}

	return -1, ErrorKeyNotInNode
}

// TODO : Remove Key
func (tree *BTree) removeKey(n *Node, key []byte) {
}

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

	if int(parent.NumKeys) >= int(tree.Order) {
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
	if int(leaf.NumKeys) >= int(tree.Order) {
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
