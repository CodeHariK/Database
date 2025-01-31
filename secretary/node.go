package secretary

// // makeNode creates a new empty node with the specified type (leaf or internal)
// func (tree *BPlusTree) makeNode(leaf bool) *Node {
// 	var children []*Node = nil
// 	if !leaf {
// 		children = make([]*Node, tree.order)
// 	}

// 	return &Node{
// 		isLeaf:   leaf,
// 		keys:     make([]Key, tree.order-1),
// 		children: children,
// 		numKeys:  0,
// 	}
// }

// // Insert adds a new key to the B+ Tree
// // Automatically handles node splitting and tree growth
// func (t *BPlusTree) Insert(key int) {
// 	if t.root == nil {
// 		t.root = makeLeaf(t.order)
// 		t.root.keys[0] = key
// 		t.root.numKeys = 1
// 		return
// 	}

// 	leaf := t.findLeaf(key)
// 	if leaf.numKeys < t.order-1 {
// 		t.insertIntoLeaf(leaf, key)
// 	} else {
// 		t.splitAndInsert(leaf, key)
// 	}
// }

// // insertIntoLeaf inserts a key into a leaf node that has space
// func (t *BPlusTree) insertIntoLeaf(leaf *Node, key int) {
// 	insertPos := 0
// 	for insertPos < leaf.numKeys && key > leaf.keys[insertPos] {
// 		insertPos++
// 	}

// 	// Shift keys to make space for new key
// 	for i := leaf.numKeys; i > insertPos; i-- {
// 		leaf.keys[i] = leaf.keys[i-1]
// 	}

// 	leaf.keys[insertPos] = key
// 	leaf.numKeys++
// }

// // splitAndInsert handles insertion into a full node by splitting it
// func (t *BPlusTree) splitAndInsert(target *Node, key int) {
// 	if target.leaf {
// 		t.splitLeafNode(target, key)
// 	} else {
// 		t.splitInternalNode(target, key)
// 	}
// }

// // splitLeafNode splits a full leaf node and distributes keys
// func (t *BPlusTree) splitLeafNode(leaf *Node, key int) {
// 	newLeaf := makeLeaf(t.order)
// 	tempKeys := make([]int, t.order)

// 	// Copy and insert new key
// 	copy(tempKeys, leaf.keys)
// 	insertPos := 0
// 	for insertPos < t.order-1 && key > tempKeys[insertPos] {
// 		insertPos++
// 	}

// 	// Make space for new key
// 	for i := t.order - 1; i > insertPos; i-- {
// 		tempKeys[i] = tempKeys[i-1]
// 	}
// 	tempKeys[insertPos] = key

// 	// Split keys between old and new leaf
// 	splitPoint := t.order / 2
// 	leaf.numKeys = splitPoint
// 	newLeaf.numKeys = t.order - splitPoint

// 	copy(leaf.keys, tempKeys[:splitPoint])
// 	copy(newLeaf.keys, tempKeys[splitPoint:])

// 	// Link leaves and update parent
// 	newLeaf.next = leaf.next
// 	leaf.next = newLeaf
// 	newLeaf.parent = leaf.parent

// 	t.insertIntoParent(leaf, newLeaf.keys[0], newLeaf)
// }

// // insertIntoParent updates parent node after splitting
// func (t *BPlusTree) insertIntoParent(left *Node, key int, right *Node) {
// 	parent := left.parent

// 	if parent == nil {
// 		// Create new root
// 		t.root = makeNode(false, t.order)
// 		t.root.keys[0] = key
// 		t.root.children[0] = left
// 		t.root.children[1] = right
// 		t.root.numKeys = 1
// 		left.parent = t.root
// 		right.parent = t.root
// 		return
// 	}

// 	// Find insertion position in parent
// 	insertPos := 0
// 	for insertPos < parent.numKeys && key > parent.keys[insertPos] {
// 		insertPos++
// 	}

// 	// Insert into parent or split if full
// 	if parent.numKeys < t.order-1 {
// 		t.insertIntoNode(parent, key, right, insertPos)
// 	} else {
// 		t.splitInternalNodeInsert(parent, key, right, insertPos)
// 	}
// }

// // insertIntoNode inserts key and child into non-full parent node
// func (t *BPlusTree) insertIntoNode(parent *Node, key int, child *Node, insertPos int) {
// 	// Shift keys and children to make space
// 	for i := parent.numKeys; i > insertPos; i-- {
// 		parent.keys[i] = parent.keys[i-1]
// 		parent.children[i+1] = parent.children[i]
// 	}

// 	parent.keys[insertPos] = key
// 	parent.children[insertPos+1] = child
// 	parent.numKeys++
// 	child.parent = parent
// }

// // Delete removes a key from the B+ Tree
// // Automatically handles rebalancing and merging
// func (t *BPlusTree) Delete(key int) {
// 	if t.root == nil {
// 		return
// 	}

// 	leaf := t.findLeaf(key)
// 	t.deleteKey(leaf, key)

// 	// Update root if empty
// 	if t.root.numKeys == 0 && !t.root.leaf {
// 		t.root = t.root.children[0]
// 		t.root.parent = nil
// 	}
// }

// // deleteKey handles key deletion from leaf or internal nodes
// func (t *BPlusTree) deleteKey(node *Node, key int) {
// 	pos := 0
// 	for pos < node.numKeys && key > node.keys[pos] {
// 		pos++
// 	}

// 	if pos < node.numKeys && key == node.keys[pos] {
// 		if node.leaf {
// 			t.deleteFromLeaf(node, pos)
// 			if node.numKeys < (t.order-1)/2 {
// 				t.rebalance(node)
// 			}
// 		} else {
// 			// Replace with predecessor and delete recursively
// 			predecessor := t.findPredecessor(node.children[pos])
// 			node.keys[pos] = predecessor
// 			t.deleteKey(node.children[pos], predecessor)
// 		}
// 	} else if !node.leaf {
// 		t.deleteKey(node.children[pos], key)
// 	}
// }

// // deleteFromLeaf removes a key from a leaf node
// func (t *BPlusTree) deleteFromLeaf(leaf *Node, pos int) {
// 	// Shift keys to fill gap
// 	for i := pos; i < leaf.numKeys-1; i++ {
// 		leaf.keys[i] = leaf.keys[i+1]
// 	}
// 	leaf.numKeys--
// }

// // rebalance ensures node maintains B+ Tree properties after deletion
// func (t *BPlusTree) rebalance(node *Node) {
// 	if node == t.root {
// 		return
// 	}

// 	parent := node.parent
// 	pos := t.getChildPosition(parent, node)

// 	// Attempt redistribution first
// 	if pos > 0 && parent.children[pos-1].numKeys > (t.order-1)/2 {
// 		t.redistributeNodes(parent.children[pos-1], node, parent, pos-1)
// 	} else if pos < parent.numKeys && parent.children[pos+1].numKeys > (t.order-1)/2 {
// 		t.redistributeNodes(node, parent.children[pos+1], parent, pos)
// 	} else {
// 		// Merge nodes if redistribution not possible
// 		if pos > 0 {
// 			t.coalesceNodes(parent.children[pos-1], node, parent, pos-1)
// 		} else {
// 			t.coalesceNodes(node, parent.children[pos+1], parent, pos)
// 		}
// 	}
// }

// // redistributeNodes balances keys between sibling nodes
// func (t *BPlusTree) redistributeNodes(left, right *Node, parent *Node, pos int) {
// 	if left.leaf {
// 		// Leaf node redistribution
// 		total := left.numKeys + right.numKeys
// 		left.numKeys = total / 2
// 		right.numKeys = total - left.numKeys

// 		// Redistribute keys
// 		temp := append(left.keys[:left.numKeys], right.keys...)
// 		copy(left.keys, temp[:left.numKeys])
// 		copy(right.keys, temp[left.numKeys:total])
// 		parent.keys[pos] = right.keys[0]
// 	} else {
// 		// Internal node redistribution
// 		borrowPos := left.numKeys - 1
// 		right.children[right.numKeys+1] = right.children[right.numKeys]
// 		for i := right.numKeys; i > 0; i-- {
// 			right.keys[i] = right.keys[i-1]
// 			right.children[i] = right.children[i-1]
// 		}
// 		right.keys[0] = parent.keys[pos]
// 		right.children[0] = left.children[borrowPos+1]
// 		right.children[0].parent = right
// 		parent.keys[pos] = left.keys[borrowPos]
// 		left.numKeys--
// 		right.numKeys++
// 	}
// }

// // coalesceNodes merges two underflow nodes
// func (t *BPlusTree) coalesceNodes(left, right *Node, parent *Node, pos int) {
// 	if left.leaf {
// 		// Merge leaf nodes
// 		copy(left.keys[left.numKeys:], right.keys[:right.numKeys])
// 		left.numKeys += right.numKeys
// 		left.next = right.next
// 	} else {
// 		// Merge internal nodes
// 		left.keys[left.numKeys] = parent.keys[pos]
// 		left.numKeys++
// 		copy(left.keys[left.numKeys:], right.keys[:right.numKeys])
// 		copy(left.children[left.numKeys:], right.children[:right.numKeys+1])
// 		for i := 0; i <= right.numKeys; i++ {
// 			right.children[i].parent = left
// 		}
// 		left.numKeys += right.numKeys
// 	}

// 	t.deleteFromInternal(parent, pos)
// 	if parent.numKeys < (t.order-1)/2 {
// 		t.rebalance(parent)
// 	}
// }

// // findLeaf locates the leaf node that would contain the specified key
// func (t *BPlusTree) findLeaf(key int) *Node {
// 	current := t.root
// 	for current != nil && !current.leaf {
// 		pos := 0
// 		for pos < current.numKeys && key >= current.keys[pos] {
// 			pos++
// 		}
// 		current = current.children[pos]
// 	}
// 	return current
// }

// // getChildPosition finds a child's index in its parent's children array
// func (t *BPlusTree) getChildPosition(parent, child *Node) int {
// 	for i := 0; i <= parent.numKeys; i++ {
// 		if parent.children[i] == child {
// 			return i
// 		}
// 	}
// 	return -1
// }

// // PrintLeaves displays all keys in leaf nodes in order
// func (t *BPlusTree) PrintLeaves() {
// 	current := t.findLeaf(-1e9) // Find leftmost leaf
// 	fmt.Print("Leaves: ")
// 	for current != nil {
// 		for i := 0; i < current.numKeys; i++ {
// 			fmt.Printf("%d ", current.keys[i])
// 		}
// 		current = current.next
// 	}
// 	fmt.Println()
// }

// // PrintTree displays the tree structure in level order
// func (t *BPlusTree) PrintTree() {
// 	if t.root == nil {
// 		fmt.Println("Empty tree")
// 		return
// 	}

// 	queue := []*Node{t.root}
// 	level := 0

// 	for len(queue) > 0 {
// 		fmt.Printf("Level %d: ", level)
// 		levelSize := len(queue)
// 		level++

// 		for i := 0; i < levelSize; i++ {
// 			node := queue[0]
// 			queue = queue[1:]

// 			// Format node keys
// 			keys := make([]string, node.numKeys)
// 			for j := 0; j < node.numKeys; j++ {
// 				keys[j] = fmt.Sprintf("%d", node.keys[j])
// 			}
// 			fmt.Printf("[%s] ", strings.Join(keys, ","))

// 			// Enqueue children
// 			if !node.leaf {
// 				for j := 0; j <= node.numKeys; j++ {
// 					if node.children[j] != nil {
// 						queue = append(queue, node.children[j])
// 					}
// 				}
// 			}
// 		}
// 		fmt.Println()
// 	}
// }

// // FindRange returns all keys in the specified [start, end] range
// func (t *BPlusTree) FindRange(start, end int) []int {
// 	result := []int{}
// 	leaf := t.findLeaf(start)

// 	for leaf != nil {
// 		for i := 0; i < leaf.numKeys; i++ {
// 			key := leaf.keys[i]
// 			if key > end {
// 				return result
// 			}
// 			if key >= start {
// 				result = append(result, key)
// 			}
// 		}
// 		leaf = leaf.next
// 	}
// 	return result
// }

// // deleteFromInternal removes a key from an internal node
// func (t *BPlusTree) deleteFromInternal(node *Node, pos int) {
// 	// Shift keys and children
// 	for i := pos; i < node.numKeys-1; i++ {
// 		node.keys[i] = node.keys[i+1]
// 	}
// 	for i := pos + 1; i < node.numKeys; i++ {
// 		node.children[i] = node.children[i+1]
// 	}

// 	node.numKeys--
// 	if node != t.root && node.numKeys < (t.order-1)/2 {
// 		t.rebalance(node)
// 	}
// }

// // splitInternalNode splits an overfull internal node
// func (t *BPlusTree) splitInternalNode(oldNode *Node, key int) {
// 	newNode := makeNode(false, t.order)
// 	tempKeys := make([]int, t.order)
// 	tempChildren := make([]*Node, t.order+1)

// 	// Copy existing data
// 	copy(tempKeys, oldNode.keys)
// 	copy(tempChildren, oldNode.children)

// 	// Find insertion point
// 	insertPos := 0
// 	for insertPos < t.order-1 && key > tempKeys[insertPos] {
// 		insertPos++
// 	}

// 	// Make space for new key
// 	for i := t.order - 1; i > insertPos; i-- {
// 		tempKeys[i] = tempKeys[i-1]
// 	}
// 	tempKeys[insertPos] = key

// 	// Split node
// 	splitPoint := t.order / 2
// 	newNode.numKeys = (t.order - 1) - splitPoint
// 	oldNode.numKeys = splitPoint

// 	// Copy data to new node
// 	copy(newNode.keys, tempKeys[splitPoint+1:t.order])
// 	copy(oldNode.keys, tempKeys[:splitPoint])
// 	copy(newNode.children, tempChildren[splitPoint+1:t.order+1])
// 	copy(oldNode.children, tempChildren[:splitPoint+1])

// 	// Update parent pointers
// 	for i := 0; i <= newNode.numKeys; i++ {
// 		if newNode.children[i] != nil {
// 			newNode.children[i].parent = newNode
// 		}
// 	}

// 	t.insertIntoParent(oldNode, tempKeys[splitPoint], newNode)
// }

// // splitInternalNodeInsert handles insertion into full internal nodes
// func (t *BPlusTree) splitInternalNodeInsert(parent *Node, key int, child *Node, insertPos int) {
// 	tempKeys := make([]int, t.order)
// 	tempChildren := make([]*Node, t.order+1)

// 	// Copy existing data
// 	copy(tempKeys, parent.keys)
// 	copy(tempChildren, parent.children)

// 	// Make space for new key
// 	for i := t.order - 1; i > insertPos; i-- {
// 		tempKeys[i] = tempKeys[i-1]
// 	}
// 	tempKeys[insertPos] = key

// 	// Make space for new child
// 	for i := t.order; i > insertPos+1; i-- {
// 		tempChildren[i] = tempChildren[i-1]
// 	}
// 	tempChildren[insertPos+1] = child

// 	// Create new node and split data
// 	newNode := makeNode(false, t.order)
// 	splitPoint := t.order / 2

// 	parent.numKeys = splitPoint
// 	copy(parent.keys, tempKeys[:splitPoint])
// 	copy(parent.children, tempChildren[:splitPoint+1])

// 	newNode.numKeys = (t.order - 1) - splitPoint
// 	copy(newNode.keys, tempKeys[splitPoint+1:t.order-1])
// 	copy(newNode.children, tempChildren[splitPoint+1:t.order])

// 	// Update parent pointers
// 	for i := 0; i <= newNode.numKeys; i++ {
// 		if newNode.children[i] != nil {
// 			newNode.children[i].parent = newNode
// 		}
// 	}

// 	t.insertIntoParent(parent, tempKeys[splitPoint], newNode)
// }

// // findPredecessor locates the largest key in the left subtree
// func (t *BPlusTree) findPredecessor(node *Node) int {
// 	for !node.leaf {
// 		node = node.children[node.numKeys]
// 	}
// 	return node.keys[node.numKeys-1]
// }

// // findSuccessor locates the smallest key in the right subtree
// func (t *BPlusTree) findSuccessor(node *Node) int {
// 	for !node.leaf {
// 		node = node.children[0]
// 	}
// 	return node.keys[0]
// }

// // Search checks if a key exists in the B+ Tree
// func (t *BPlusTree) Search(key int) bool {
// 	if t.root == nil {
// 		return false
// 	}

// 	leaf := t.findLeaf(key)
// 	for i := 0; i < leaf.numKeys; i++ {
// 		if leaf.keys[i] == key {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (t *BPlusTree) WriteNode(n *Node) error {
// 	var file *os.File
// 	var offset int64

// 	if n.IsLeaf {
// 		file = t.leafFile
// 		offset = n.ID * ChunkSize
// 	} else {
// 		file = t.internalFile
// 		offset = n.ID * ChunkSize
// 	}

// 	var data []byte
// 	if n.IsLeaf {
// 		data = n.serializeLeaf()
// 	} else {
// 		data = n.serializeInternal()
// 	}

// 	_, err := file.WriteAt(data, offset)
// 	return err
// }

// func (t *BPlusTree) ReadNode(id int64, isLeaf bool) (*Node, error) {
// 	var file *os.File
// 	var offset int64

// 	if isLeaf {
// 		file = t.leafFile
// 	} else {
// 		file = t.internalFile
// 	}
// 	offset = id * ChunkSize

// 	buf := make([]byte, ChunkSize)
// 	_, err := file.ReadAt(buf, offset)
// 	if err != nil {
// 		return nil, err
// 	}

// 	n := &Node{
// 		ID:       id,
// 		IsLeaf:   buf[IsLeafOffset] == 1,
// 		NumKeys:  int(binary.LittleEndian.Uint32(buf[NumKeysOffset:])),
// 		ParentID: int64(binary.LittleEndian.Uint64(buf[ParentIDOffset:])),
// 	}

// 	if n.IsLeaf {
// 		n.NextID = int64(binary.LittleEndian.Uint64(buf[LeafNextIDOffset:]))
// 		n.Keys = make([]int32, MaxOrder-1)
// 		n.RecordIDs = make([]int64, MaxOrder-1)

// 		for i := 0; i < MaxOrder-1; i++ {
// 			n.Keys[i] = int32(binary.LittleEndian.Uint32(buf[LeafKeysStart+i*4:]))
// 			n.RecordIDs[i] = int64(binary.LittleEndian.Uint64(buf[LeafRecordsStart+i*8:]))
// 		}
// 	} else {
// 		n.Children = make([]int64, MaxOrder)
// 		n.Keys = make([]int32, MaxOrder-1)

// 		for i := 0; i < MaxOrder-1; i++ {
// 			n.Keys[i] = int32(binary.LittleEndian.Uint32(buf[InternalKeysStart+i*4:]))
// 		}

// 		for i := 0; i < MaxOrder; i++ {
// 			n.Children[i] = int64(binary.LittleEndian.Uint64(buf[InternalChildStart+i*8:]))
// 		}
// 	}

// 	return n, nil
// }
