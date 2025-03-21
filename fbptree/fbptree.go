package fbptree

import (
	"fmt"
	"math"
	"os"
)

const defaultOrder = 500

const (
	maxKeySize   = math.MaxUint16
	maxValueSize = math.MaxUint16
	maxTreeSize  = math.MaxUint32
)

// the limit for the  B+ tree order, must be less than math.MaxUint16
const maxOrder = 1000

// FBPTree represents B+ tree store in the file.
type FBPTree struct {
	order int

	storage *storage

	metadata *treeMetadata

	// minimum allowed number of keys in the tree ceil(order/2)-1
	minKeyNum int
}

// Order option specifies the order of the B+ tree, between 3 and 1000.
func Order(order int) func(*config) error {
	return func(c *config) error {
		if order < 3 {
			return fmt.Errorf("order must be >= 3")
		}

		if order > maxOrder {
			return fmt.Errorf("order must be <= %d", maxOrder)
		}

		c.order = uint16(order)

		return nil
	}
}

// PageSize option specifies the page size for the B+ tree file.
func PageSize(pageSize int) func(*config) error {
	return func(t *config) error {
		if pageSize < minPageSize {
			return fmt.Errorf("page size must be greater than or equal to %d", minPageSize)
		}

		if pageSize > maxPageSize {
			return fmt.Errorf("page size must not be greater than %d", maxPageSize)
		}

		t.pageSize = uint16(pageSize)

		return nil
	}
}

// Open opens an existent B+ tree or creates a new file.
func Open(path string, options ...func(*config) error) (*FBPTree, error) {
	defaultPageSize := os.Getpagesize()
	if defaultPageSize > maxPageSize {
		defaultPageSize = maxPageSize
	}

	cfg := &config{pageSize: uint16(defaultPageSize), order: defaultOrder}
	for _, option := range options {
		err := option(cfg)
		if err != nil {
			return nil, err
		}
	}

	storage, err := newStorage(path, cfg.pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the storage: %w", err)
	}

	metadata, err := storage.loadMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load the metadata: %w", err)
	}

	if metadata != nil && metadata.order != cfg.order {
		return nil, fmt.Errorf("the tree was created with %d order, but the new order value is given %d", metadata.order, cfg.order)
	}

	minKeyNum := ceil(int(cfg.order), 2) - 1

	return &FBPTree{storage: storage, order: int(cfg.order), metadata: metadata, minKeyNum: minKeyNum}, nil
}

// Get return the value by the key. Returns true if the
// key exists.
func (t *FBPTree) Get(key []byte) ([]byte, bool, error) {
	if t.metadata == nil {
		return nil, false, nil
	}

	leaf, err := t.findLeaf(key)
	if err != nil {
		return nil, false, fmt.Errorf("failed to find leaf: %w", err)
	}

	for i := 0; i < leaf.keyNum; i++ {
		if compare(key, leaf.keys[i]) == 0 {
			return leaf.pointers[i].asValue(), true, nil
		}
	}

	return nil, false, nil
}

// findLeaf finds a leaf that might contain the key.
func (t *FBPTree) findLeaf(key []byte) (*node, error) {
	root, err := t.storage.loadNodeByID(t.metadata.rootID)
	if err != nil {
		return nil, fmt.Errorf("failed to load root node: %w", err)
	}

	current := root
	for !current.leaf {
		position := 0
		for position < current.keyNum {
			if less(key, current.keys[position]) {
				break
			} else {
				position += 1
			}
		}

		nextID := current.pointers[position].asNodeID()
		nextNode, err := t.storage.loadNodeByID(nextID)
		if err != nil {
			return nil, fmt.Errorf("failed to load next node %d: %w", nextID, err)
		}

		current = nextNode
	}

	return current, nil
}

// Put puts the key and the value into the tree. Returns true if the
// key already exists and anyway overwrites it.
func (t *FBPTree) Put(key, value []byte) ([]byte, bool, error) {
	if len(key) > maxKeySize {
		return nil, false, fmt.Errorf("maximum key size is %d, but received %d", maxKeySize, len(key))
	} else if len(value) > maxValueSize {
		return nil, false, fmt.Errorf("maximum value size is %d, but received %d", maxValueSize, len(value))
	} else if t.metadata != nil && t.metadata.size >= maxTreeSize {
		return nil, false, fmt.Errorf("maximum tree size is reached: %d", maxTreeSize)
	}

	if t.metadata == nil {
		err := t.initializeRoot(key, value)
		if err != nil {
			return nil, false, fmt.Errorf("failed to initialize root: %w", err)
		}

		return nil, false, nil
	}

	leaf, err := t.findLeaf(key)
	if err != nil {
		return nil, false, fmt.Errorf("failed to find leaf: %w", err)
	}

	oldValue, overridden, err := t.putIntoLeaf(leaf, key, value)
	if err != nil {
		return nil, false, fmt.Errorf("failed to put into the leaf %d: %w", leaf.id, err)
	}

	return oldValue, overridden, nil
}

// initializeRoot initializes root in the empty tree.
func (t *FBPTree) initializeRoot(key, value []byte) error {
	newNodeID, err := t.storage.newNode()
	if err != nil {
		return fmt.Errorf("failed to instantiate new node: %w", err)
	}

	// new tree
	keys := make([][]byte, t.order-1)
	keys[0] = copyBytes(key)

	pointers := make([]*pointer, t.order)
	pointers[0] = &pointer{value}

	rootNode := &node{
		id:       newNodeID,
		leaf:     true,
		parentID: 0,
		keys:     keys,
		keyNum:   1,
		pointers: pointers,
	}

	err = t.storage.updateNodeByID(newNodeID, rootNode)
	if err != nil {
		return fmt.Errorf("failed to store root node: %w", err)
	}

	err = t.updateMetadata(newNodeID, newNodeID, 1)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

func (t *FBPTree) updateMetadata(rootID, leftmostID, size uint32) error {
	if t.metadata == nil {
		// initialization
		t.metadata = new(treeMetadata)
		t.metadata.order = uint16(t.order)
	}

	t.metadata.rootID = rootID
	t.metadata.leftmostID = leftmostID
	t.metadata.size = size

	err := t.storage.updateMetadata(t.metadata)
	if err != nil {
		return fmt.Errorf("failed to store metadata: %w", err)
	}

	return nil
}

func (t *FBPTree) deleteMetadata() error {
	t.metadata = nil

	err := t.storage.deleteMetadata()
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

// putIntoNewRoot creates new root, inserts left and right entries
// and updates the tree.
func (t *FBPTree) putIntoNewRoot(key []byte, l, r *node) error {
	newNodeID, err := t.storage.newNode()
	if err != nil {
		return fmt.Errorf("failed to instantiate new node: %w", err)
	}

	// new root
	newRoot := &node{
		id:       newNodeID,
		leaf:     false,
		keys:     make([][]byte, t.order-1),
		pointers: make([]*pointer, t.order),
		parentID: 0,
		keyNum:   1, // we are going to put just one key
	}

	newRoot.keys[0] = key
	newRoot.pointers[0] = &pointer{l.id}
	newRoot.pointers[1] = &pointer{r.id}

	err = t.storage.updateNodeByID(newNodeID, newRoot)
	if err != nil {
		return fmt.Errorf("failed to update node by ID %d: %w", newNodeID, err)
	}

	l.parentID = newNodeID
	err = t.storage.updateNodeByID(l.id, l)
	if err != nil {
		return fmt.Errorf("failed to update left node %d: %w", l.id, err)
	}

	r.parentID = newNodeID
	err = t.storage.updateNodeByID(r.id, r)
	if err != nil {
		return fmt.Errorf("failed to update right node %d: %w", r.id, err)
	}

	err = t.updateRootID(newNodeID)
	if err != nil {
		return fmt.Errorf("failed to update root ID to %d: %w", newNodeID, err)
	}

	return nil
}

func (t *FBPTree) updateSize(size uint32) error {
	return t.updateMetadata(t.metadata.rootID, t.metadata.leftmostID, size)
}

func (t *FBPTree) updateRootID(rootID uint32) error {
	var leftmostID uint32
	if t.metadata != nil {
		leftmostID = t.metadata.leftmostID
	}

	return t.updateMetadata(rootID, leftmostID, t.metadata.size)
}

// putIntoLeaf puts key and value into the node.
func (t *FBPTree) putIntoLeaf(n *node, k, v []byte) ([]byte, bool, error) {
	insertPos := 0
	for insertPos < n.keyNum {
		cmp := compare(k, n.keys[insertPos])
		if cmp == 0 {
			// found the exact match
			oldValue := n.pointers[insertPos].overrideValue(v)

			err := t.storage.updateNodeByID(n.id, n)
			if err != nil {
				return nil, false, fmt.Errorf("failed to update the node %d: %w", n.id, err)
			}

			return oldValue, true, nil
		} else if cmp < 0 {
			// found the insert position,
			// can break the loop
			break
		}

		insertPos++
	}

	// if we did not find the same key, we continue to insert
	if n.keyNum < len(n.keys) {
		// if the node is not full

		// shift the keys and pointers
		for j := n.keyNum; j > insertPos; j-- {
			n.keys[j] = n.keys[j-1]
			n.pointers[j] = n.pointers[j-1]
		}

		// insert
		n.keys[insertPos] = k
		n.pointers[insertPos] = &pointer{v}
		// and update key num
		n.keyNum++

		err := t.storage.updateNodeByID(n.id, n)
		if err != nil {
			return nil, false, fmt.Errorf("failed to update the node %d: %w", n.id, err)
		}
	} else {
		// if the node is full
		var parentNode *node
		if n.parentID != 0 {
			p, err := t.storage.loadNodeByID(n.parentID)
			if err != nil {
				return nil, false, fmt.Errorf("failed to load parent node %d: %w", n.parentID, err)
			}

			parentNode = p
		}
		parent := parentNode

		left, right, err := t.putIntoLeafAndSplit(n, insertPos, k, v)
		if err != nil {
			return nil, false, fmt.Errorf("failed to split the node %d: %w", n.id, err)
		}

		insertKey := right.keys[0]
		for left != nil && right != nil {
			if parent == nil {
				t.putIntoNewRoot(insertKey, left, right)
				break
			} else {
				if parent.keyNum < len(parent.keys) {
					// if the parent is not full
					err := t.putIntoParent(parent, insertKey, left, right)
					if err != nil {
						return nil, false, fmt.Errorf("failed to put into the parent: %w", err)
					}

					break
				} else {
					// if the parent is full
					// split parent, insert into the new parent and continue
					insertKey, left, right, err = t.putIntoParentAndSplit(parent, insertKey, left, right)
					if err != nil {
						return nil, false, fmt.Errorf("failed to put into the parent and split: %w", err)
					}
				}
			}

			var parentParentNode *node
			if parent.parentID != 0 {
				p, err := t.storage.loadNodeByID(parent.parentID)
				if err != nil {
					return nil, false, fmt.Errorf("failed to load the parent of the parent node %d: %w", parent.parentID, err)
				}

				parentParentNode = p
			}

			parent = parentParentNode
		}
	}

	t.metadata.size++
	err := t.updateSize(t.metadata.size)
	if err != nil {
		return nil, false, fmt.Errorf("failed to update the tree size to %d: %w", t.metadata.size, err)
	}

	return nil, false, nil
}

// putIntoParent puts the node into the parent and update the left and the right
// pointers.
func (t *FBPTree) putIntoParent(parent *node, k []byte, l, r *node) error {
	insertPos := 0
	for insertPos < parent.keyNum {
		if less(k, parent.keys[insertPos]) {
			// found the insert position,
			// can break the loop
			break
		}

		insertPos++
	}

	// shift the keys and pointers
	parent.pointers[parent.keyNum+1] = parent.pointers[parent.keyNum]
	for j := parent.keyNum; j > insertPos; j-- {
		parent.keys[j] = parent.keys[j-1]
		parent.pointers[j] = parent.pointers[j-1]
	}

	// insert
	parent.keys[insertPos] = k
	parent.pointers[insertPos] = &pointer{l.id}
	parent.pointers[insertPos+1] = &pointer{r.id}
	// and update key num
	parent.keyNum++

	err := t.storage.updateNodeByID(parent.id, parent)
	if err != nil {
		return fmt.Errorf("failed to update parent node %d: %w", parent.id, err)
	}

	l.parentID = parent.id
	err = t.storage.updateNodeByID(l.id, l)
	if err != nil {
		return fmt.Errorf("failed to update left node %d: %w", l.id, err)
	}

	r.parentID = parent.id
	err = t.storage.updateNodeByID(r.id, r)
	if err != nil {
		return fmt.Errorf("failed to update right node %d: %w", r.id, err)
	}

	return nil
}

// putIntoParentAndSplit puts key in the parent, splits the node and returns the splitten
// nodes with all fixed pointers.
func (t *FBPTree) putIntoParentAndSplit(parent *node, k []byte, l, r *node) ([]byte, *node, *node, error) {
	insertPos := 0
	for insertPos < parent.keyNum {
		if less(k, parent.keys[insertPos]) {
			// found the insert position,
			// can break the loop
			break
		}

		insertPos++
	}

	newNodeID, err := t.storage.newNode()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to instantiate new node: %w", err)
	}

	right := &node{
		id:       newNodeID,
		leaf:     false,
		keys:     make([][]byte, t.order-1),
		keyNum:   0,
		pointers: make([]*pointer, t.order),
		parentID: 0,
	}

	middlePos := ceil(len(parent.keys), 2)
	copyFrom := middlePos
	if insertPos < middlePos {
		// since the elements will be shifted
		copyFrom -= 1
	}

	copy(right.keys, parent.keys[copyFrom:])
	copy(right.pointers, parent.pointers[copyFrom:])
	// copy the pointer to the next node
	right.keyNum = len(right.keys) - copyFrom

	// the given node becomes the left node
	left := parent
	left.keyNum = copyFrom
	// clean up keys and pointers
	for i := len(left.keys) - 1; i >= copyFrom; i-- {
		left.keys[i] = nil
		left.pointers[i+1] = nil
	}

	insertNode := left
	if insertPos >= middlePos {
		insertNode = right
		insertPos -= middlePos
	}

	// insert into the node
	insertNode.pointers[insertNode.keyNum+1] = insertNode.pointers[insertNode.keyNum]
	for j := insertNode.keyNum; j > insertPos; j-- {
		insertNode.keys[j] = insertNode.keys[j-1]
		insertNode.pointers[j] = insertNode.pointers[j-1]
	}

	insertNode.keys[insertPos] = k
	insertNode.pointers[insertPos] = &pointer{l.id}
	insertNode.pointers[insertPos+1] = &pointer{r.id}
	insertNode.keyNum++

	l.parentID = insertNode.id
	err = t.storage.updateNodeByID(l.id, l)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to update the l node %d: %w", parent.id, err)
	}

	r.parentID = insertNode.id
	err = t.storage.updateNodeByID(r.id, r)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to update the r node %d: %w", right.id, err)
	}

	middleKey := right.keys[0]

	// clean up the right node
	for i := 1; i < right.keyNum; i++ {
		right.keys[i-1] = right.keys[i]
		right.pointers[i-1] = right.pointers[i]
	}
	right.pointers[right.keyNum-1] = right.pointers[right.keyNum]
	right.pointers[right.keyNum] = nil
	right.keys[right.keyNum-1] = nil
	right.keyNum--

	// update the pointers
	for _, p := range left.pointers {
		if p != nil {
			nodeID := p.asNodeID()
			node, err := t.storage.loadNodeByID(nodeID)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to load the node by id %d: %w", nodeID, err)
			}

			if node.parentID == left.id {
				continue
			}

			node.parentID = left.id
			err = t.storage.updateNodeByID(node.id, node)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to update node by id %d: %w", node.id, err)
			}
		}
	}

	for _, p := range right.pointers {
		if p != nil {
			nodeID := p.asNodeID()
			node, err := t.storage.loadNodeByID(nodeID)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to load the node by id %d: %w", nodeID, err)
			}

			if node.parentID == right.id {
				continue
			}

			node.parentID = right.id
			err = t.storage.updateNodeByID(node.id, node)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to update node by id %d: %w", node.id, err)
			}
		}
	}

	err = t.storage.updateNodeByID(parent.id, parent)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to update the right node %d: %w", right.id, err)
	}
	err = t.storage.updateNodeByID(right.id, right)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to update the right node %d: %w", right.id, err)
	}
	err = t.storage.updateNodeByID(left.id, left)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to update the left node %d: %w", left.id, err)
	}

	return middleKey, left, right, nil
}

// putIntoLeafAndSplit puts the new key and splits the node into the left and right nodes
// and returns the left and the right nodes.
// The given node becomes left node.
// The tree is right-biased, so the first element in
// the right node is the "middle" key.
func (t *FBPTree) putIntoLeafAndSplit(n *node, insertPos int, k, v []byte) (*node, *node, error) {
	newNodeID, err := t.storage.newNode()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to instantiate new node: %w", err)
	}

	right := &node{
		id:       newNodeID,
		leaf:     true,
		keys:     make([][]byte, t.order-1),
		keyNum:   0,
		pointers: make([]*pointer, t.order),
		parentID: 0,
	}

	middlePos := ceil(len(n.keys), 2)
	copyFrom := middlePos
	if insertPos < middlePos {
		// since the elements will be shifted
		copyFrom -= 1
	}

	copy(right.keys, n.keys[copyFrom:])
	copy(right.pointers, n.pointers[copyFrom:len(n.pointers)-1])

	// copy the pointer to the next node
	right.setNext(n.next())
	right.keyNum = len(right.keys) - copyFrom

	// the given node becomes the left node
	left := n
	left.parentID = 0
	left.keyNum = copyFrom
	// clean up keys and pointers
	for i := len(left.keys) - 1; i >= copyFrom; i-- {
		left.keys[i] = nil
		left.pointers[i] = nil
	}
	left.setNext(&pointer{right.id})

	insertNode := left
	if insertPos >= middlePos {
		insertNode = right
		// normalize insert position
		insertPos -= middlePos
	}

	// insert into the node
	insertNode.insertAt(insertPos, k, insertPos, &pointer{v})

	err = t.storage.updateNodeByID(right.id, right)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update the right node %d: %w", right.id, err)
	}

	err = t.storage.updateNodeByID(left.id, left)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update the left node %d: %w", left.id, err)
	}

	return left, right, nil
}

// Delete deletes the value by the key. Returns true if the
// key exists.
func (t *FBPTree) Delete(key []byte) ([]byte, bool, error) {
	if t.metadata == nil {
		return nil, false, nil
	}

	leaf, err := t.findLeaf(key)
	if err != nil {
		return nil, false, fmt.Errorf("failed to find the leaf: %w", err)
	}

	value, deleted, err := t.deleteAtLeafAndRebalance(leaf, key)
	if err != nil {
		return nil, false, fmt.Errorf("failed to delete and rebalance: %w", err)
	}

	if !deleted {
		return nil, false, nil
	}

	if t.metadata != nil {
		t.metadata.size--
		err = t.updateSize(t.metadata.size)
		if err != nil {
			return nil, false, fmt.Errorf("failed to update the tree size to %d: %w", t.metadata.size, err)
		}
	}

	return value, true, nil
}

// deleteAtLeafAndRebalance deletes the key from the given node and rebalances it.
func (t *FBPTree) deleteAtLeafAndRebalance(n *node, key []byte) ([]byte, bool, error) {
	keyPos := n.keyPosition(key)
	if keyPos == -1 {
		return nil, false, nil
	}

	value := n.pointers[keyPos].asValue()
	n.deleteAt(keyPos, keyPos)
	err := t.storage.updateNodeByID(n.id, n)
	if err != nil {
		return nil, false, fmt.Errorf("failed to update the node by id %d: %w", n.id, err)
	}

	if n.parentID == 0 {
		if n.keyNum == 0 {
			// remove the root (as leaf)
			err := t.storage.deleteNodeByID(n.id)
			if err != nil {
				return nil, false, fmt.Errorf("failed to delete the node by id %d: %w", n.id, err)
			}

			err = t.deleteMetadata()
			if err != nil {
				return nil, false, fmt.Errorf("failed to delete the metadata: %w", err)
			}
		} else {
			// update the root
			err := t.storage.updateNodeByID(n.id, n)
			if err != nil {
				return nil, false, fmt.Errorf("failed to update the node by id %d: %w", n.id, err)
			}
		}

		return value, true, nil
	}

	if n.keyNum < t.minKeyNum {
		err := t.rebalanceFromLeafNode(n)
		if err != nil {
			return nil, false, fmt.Errorf("failed to rebalance from the leaf node: %w", err)
		}
	}

	err = t.removeFromIndex(key)
	if err != nil {
		return nil, false, fmt.Errorf("failed to remove the key from the index: %w", err)
	}

	return value, true, nil
}

// removeFromIndex searches the key in the index (internal nodes and if finds it changes to
// the leftmost key in the right subtree.
func (t *FBPTree) removeFromIndex(key []byte) error {
	root, err := t.storage.loadNodeByID(t.metadata.rootID)
	if err != nil {
		return fmt.Errorf("failed to load the root node %d: %w", t.metadata.rootID, err)
	}

	current := root
	for !current.leaf {
		// until the leaf is reached

		position := 0
		for position < current.keyNum {
			cmp := compare(key, current.keys[position])
			if cmp < 0 {
				break
			} else if cmp > 0 {
				position += 1
			} else if cmp == 0 {
				// the key is found in the index
				// take the right sub-tree and find the leftmost key
				// and update the key
				nodeID := current.pointers[position+1].asNodeID()
				leftmostKey, err := t.findLeftmostKey(nodeID)
				if err != nil {
					return fmt.Errorf("failed to find the leftmost key for %d: %w", nodeID, err)
				}
				current.keys[position] = leftmostKey

				err = t.storage.updateNodeByID(current.id, current)
				if err != nil {
					return fmt.Errorf("failed to update the node %d: %w", current.id, err)
				}
			}
		}

		nextNodeID := current.pointers[position].asNodeID()
		nextNode, err := t.storage.loadNodeByID(nextNodeID)
		if err != nil {
			return fmt.Errorf("failed to load the next node node %d: %w", nextNodeID, err)
		}

		current = nextNode
	}

	return nil
}

// findLeftmostKey returns the leftmost key for the node.
func (t *FBPTree) findLeftmostKey(nodeID uint32) ([]byte, error) {
	node, err := t.storage.loadNodeByID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load the node by id %d: %w", nodeID, err)
	}

	current := node
	for !current.leaf {
		nextID := current.pointers[0].asNodeID()
		nextNode, err := t.storage.loadNodeByID(nextID)
		if err != nil {
			return nil, fmt.Errorf("failed to load the next node by id %d: %w", nextID, err)
		}

		current = nextNode
	}

	return current.keys[0], nil
}

// rebalanceFromLeafNode starts rebalancing the tree from the leaf node.
func (t *FBPTree) rebalanceFromLeafNode(n *node) error {
	parent, err := t.storage.loadNodeByID(n.parentID)
	if err != nil {
		return fmt.Errorf("failed to load the parent node by id %d: %w", n.parentID, err)
	}

	pointerPositionInParent := parent.pointerPositionOf(n)
	keyPositionInParent := pointerPositionInParent - 1
	if keyPositionInParent < 0 {
		keyPositionInParent = 0
	}

	// trying to borrow for the leaf from any sibling

	// check left sibling
	leftSiblingPosition := pointerPositionInParent - 1
	var leftSibling *node
	if leftSiblingPosition >= 0 {
		// if left sibling exists
		leftSiblingID := parent.pointers[leftSiblingPosition].asNodeID()
		ls, err := t.storage.loadNodeByID(leftSiblingID)
		if err != nil {
			return fmt.Errorf("failed to load the left sibling node by id %d: %w", leftSiblingID, err)
		}
		leftSibling = ls

		if leftSibling.keyNum > t.minKeyNum {
			// borrow from the left sibling
			n.insertAt(0, leftSibling.keys[leftSibling.keyNum-1], 0, leftSibling.pointers[leftSibling.keyNum-1])
			leftSibling.deleteAt(leftSibling.keyNum-1, leftSibling.keyNum-1)
			parent.keys[keyPositionInParent] = n.keys[0]

			err = t.storage.updateNodeByID(n.id, n)
			if err != nil {
				return fmt.Errorf("failed to update the node by id %d: %w", n.id, err)
			}
			err = t.storage.updateNodeByID(leftSibling.id, leftSibling)
			if err != nil {
				return fmt.Errorf("failed to update the left sibling node by id %d: %w", leftSibling.id, err)
			}
			err = t.storage.updateNodeByID(parent.id, parent)
			if err != nil {
				return fmt.Errorf("failed to update the parent node by id %d: %w", parent.id, err)
			}

			return nil
		}
	}

	rightSiblingPosition := pointerPositionInParent + 1
	var rightSibling *node
	if rightSiblingPosition < parent.keyNum+1 {
		// if right sibling exists
		rightSiblingID := parent.pointers[rightSiblingPosition].asNodeID()
		rs, err := t.storage.loadNodeByID(rightSiblingID)
		if err != nil {
			return fmt.Errorf("failed to load the right sibling node by id %d: %w", rightSiblingID, err)
		}
		rightSibling = rs

		if rightSibling.keyNum > t.minKeyNum {
			// borrow from the right sibling
			n.append(rightSibling.keys[0], rightSibling.pointers[0], t.storage)
			rightSibling.deleteAt(0, 0)
			parent.keys[rightSiblingPosition-1] = rightSibling.keys[0]

			err := t.storage.updateNodeByID(n.id, n)
			if err != nil {
				return fmt.Errorf("failed to update the node by id %d: %w", n.id, err)
			}
			err = t.storage.updateNodeByID(rightSibling.id, rightSibling)
			if err != nil {
				return fmt.Errorf("failed to update the right sibling node by id %d: %w", rightSibling.id, err)
			}
			err = t.storage.updateNodeByID(parent.id, parent)
			if err != nil {
				return fmt.Errorf("failed to update the parent node by id %d: %w", parent.id, err)
			}

			return nil
		}
	}

	// if we could borrow, we would borrow
	// so, we just take the first available sibling and merge with it
	// and the remove the navigator key and appropriate pointer

	// merge nodes and remove the "navigator" key and appropriate
	if leftSibling != nil {
		err := leftSibling.copyFromRight(n, t.storage)
		if err != nil {
			return fmt.Errorf("failed to copy to the left sibling %d: %w", rightSibling.id, err)
		}
		parent.deleteAt(keyPositionInParent, pointerPositionInParent)

		err = t.storage.updateNodeByID(leftSibling.id, leftSibling)
		if err != nil {
			return fmt.Errorf("failed to update the left sibling node by id %d: %w", parent.id, err)
		}
		err = t.storage.updateNodeByID(parent.id, parent)
		if err != nil {
			return fmt.Errorf("failed to update the parent node by id %d: %w", parent.id, err)
		}
	} else if rightSibling != nil {
		err := n.copyFromRight(rightSibling, t.storage)
		if err != nil {
			return fmt.Errorf("failed to copy from the right sibling %d: %w", rightSibling.id, err)
		}
		parent.deleteAt(keyPositionInParent, rightSiblingPosition)

		err = t.storage.updateNodeByID(n.id, n)
		if err != nil {
			return fmt.Errorf("failed to update the node by id %d: %w", n.id, err)
		}
		err = t.storage.updateNodeByID(parent.id, parent)
		if err != nil {
			return fmt.Errorf("failed to update the parent node by id %d: %w", parent.id, err)
		}
	}

	err = t.rebalanceParentNode(parent)
	if err != nil {
		return fmt.Errorf("failed to rebalance the parent node %d: %w", parent.id, err)
	}

	return nil
}

// rebalanceInternalNode rebalances the tree from the internal node. It expects that
func (t *FBPTree) rebalanceParentNode(n *node) error {
	if n.parentID == 0 {
		if n.keyNum == 0 {
			rootID := n.pointers[0].asNodeID()

			root, err := t.storage.loadNodeByID(rootID)
			if err != nil {
				return fmt.Errorf("failed to load the root node by id %d", rootID)
			}

			root.parentID = 0

			err = t.storage.updateNodeByID(rootID, root)
			if err != nil {
				return fmt.Errorf("failed to update the root node %d: %w", rootID, err)
			}

			err = t.updateRootID(rootID)
			if err != nil {
				return fmt.Errorf("failed to update the root id to %d", rootID)
			}
		}

		return nil
	}

	if n.keyNum >= t.minKeyNum {
		// balanced
		return nil
	}

	parent, err := t.storage.loadNodeByID(n.parentID)
	if err != nil {
		return fmt.Errorf("failed to load parent node %d: %w", n.parentID, err)
	}

	pointerPositionInParent := parent.pointerPositionOf(n)
	keyPositionInParent := pointerPositionInParent - 1
	if keyPositionInParent < 0 {
		keyPositionInParent = 0
	}

	// trying to borrow for the internal node from any sibling

	// check left sibling
	leftSiblingPosition := pointerPositionInParent - 1
	var leftSibling *node
	if leftSiblingPosition >= 0 {
		leftSiblingID := parent.pointers[leftSiblingPosition].asNodeID()
		// if left sibling exists
		ls, err := t.storage.loadNodeByID(leftSiblingID)
		if err != nil {
			return fmt.Errorf("failed to load the left sibling %d: %w", leftSiblingID, err)
		}
		leftSibling = ls

		if leftSibling.keyNum > t.minKeyNum {
			splitKey := parent.keys[keyPositionInParent]

			// borrow from the left sibling
			childID := leftSibling.pointers[leftSibling.keyNum].asNodeID()
			child, err := t.storage.loadNodeByID(childID)
			if err != nil {
				return fmt.Errorf("failed to load the child node %d for the left sibling %d: %w", childID, leftSiblingID, err)
			}

			child.parentID = n.id

			err = t.storage.updateNodeByID(child.id, child)
			if err != nil {
				return fmt.Errorf("failed to update the child node %d for the left sibling %d: %w", childID, leftSiblingID, err)
			}

			n.insertAt(0, splitKey, 0, leftSibling.pointers[leftSibling.keyNum])

			parent.keys[keyPositionInParent] = leftSibling.keys[leftSibling.keyNum-1]
			leftSibling.deleteAt(leftSibling.keyNum-1, leftSibling.keyNum)

			err = t.storage.updateNodeByID(n.id, n)
			if err != nil {
				return fmt.Errorf("failed to update the node by id %d: %w", n.id, err)
			}

			err = t.storage.updateNodeByID(parent.id, parent)
			if err != nil {
				return fmt.Errorf("failed to update the parent node %d: %w", parent.id, err)
			}
			err = t.storage.updateNodeByID(leftSibling.id, leftSibling)
			if err != nil {
				return fmt.Errorf("failed to update the left sibling %d: %w", leftSibling.id, err)
			}

			return nil
		}
	}

	rightSiblingPosition := pointerPositionInParent + 1
	var rightSibling *node
	if rightSiblingPosition < parent.keyNum+1 {
		// if right sibling exists
		rightSiblingID := parent.pointers[rightSiblingPosition].asNodeID()
		rs, err := t.storage.loadNodeByID(rightSiblingID)
		if err != nil {
			return fmt.Errorf("failed to load the right sibling id %d: %w", rightSiblingID, err)
		}
		rightSibling = rs

		if rightSibling.keyNum > t.minKeyNum {
			splitKeyPosition := rightSiblingPosition - 1
			splitKey := parent.keys[splitKeyPosition]

			// borrow from the right sibling
			err := n.append(splitKey, rightSibling.pointers[0], t.storage)
			if err != nil {
				return fmt.Errorf("failed to append to node %d: %w", n.id, err)
			}

			parent.keys[splitKeyPosition] = rightSibling.keys[0]
			rightSibling.deleteAt(0, 0)

			err = t.storage.updateNodeByID(n.id, n)
			if err != nil {
				return fmt.Errorf("failed to update the node by id %d: %w", n.id, err)
			}
			err = t.storage.updateNodeByID(parent.id, parent)
			if err != nil {
				return fmt.Errorf("failed to update the parent node %d: %w", parent.id, err)
			}
			err = t.storage.updateNodeByID(rightSibling.id, rightSibling)
			if err != nil {
				return fmt.Errorf("failed to update the right sibling %d: %w", rightSibling.id, err)
			}

			return nil
		}
	}

	// if we could borrow, we would borrow
	// so, we just take the first available sibling and merge with it
	if leftSibling != nil {
		splitKey := parent.keys[keyPositionInParent]

		// incorporate the split key from parent for the merging
		leftSibling.keys[leftSibling.keyNum] = splitKey
		leftSibling.keyNum++

		err := leftSibling.copyFromRight(n, t.storage)
		if err != nil {
			return fmt.Errorf("failed to copy from to left sibling %d: %w", leftSibling.id, err)
		}
		err = t.storage.updateNodeByID(leftSibling.id, leftSibling)
		if err != nil {
			return fmt.Errorf("failed to update the left sibling by id %d: %w", leftSibling.id, err)
		}

		parent.deleteAt(keyPositionInParent, pointerPositionInParent)
		err = t.storage.updateNodeByID(parent.id, parent)
		if err != nil {
			return fmt.Errorf("failed to update the parent node %d: %w", parent.id, err)
		}
	} else if rightSibling != nil {
		splitKey := parent.keys[keyPositionInParent]

		n.keys[n.keyNum] = splitKey
		n.keyNum++

		err = n.copyFromRight(rightSibling, t.storage)
		if err != nil {
			return fmt.Errorf("failed to copy from the right sibling %d: %w", rightSibling.id, err)
		}

		err = t.storage.updateNodeByID(n.id, n)
		if err != nil {
			return fmt.Errorf("failed to update the node by id %d: %w", n.id, err)
		}

		parent.deleteAt(keyPositionInParent, rightSiblingPosition)
		err = t.storage.updateNodeByID(parent.id, parent)
		if err != nil {
			return fmt.Errorf("failed to update the parent node %d: %w", parent.id, err)
		}
	}

	err = t.rebalanceParentNode(parent)
	if err != nil {
		return fmt.Errorf("failed to rebalance the parent node %d: %w", parent.id, err)
	}

	return nil
}

// ForEach traverses tree in ascending key order.
func (t *FBPTree) ForEach(action func(key []byte, value []byte)) error {
	it, err := t.Iterator()
	if err != nil {
		return fmt.Errorf("failed to initialize iterator: %w", err)
	}

	for it := it; it.HasNext(); {
		key, value, err := it.Next()
		if err != nil {
			return fmt.Errorf("failed to advance to the next element: %w", err)
		}

		action(key, value)
	}

	return nil
}

// Size return the size of the tree.
func (t *FBPTree) Size() int {
	if t.metadata != nil {
		return int(t.metadata.size)
	}

	return 0
}

// Close closes the tree and free the underlying resources.
func (t *FBPTree) Close() error {
	if err := t.storage.close(); err != nil {
		return fmt.Errorf("failed to close the storage: %w", err)
	}

	return nil
}
