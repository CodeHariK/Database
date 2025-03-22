* ```go

* convert byte[] to records and node
* traverse entire tree and store it in disk
* Store entire node in same page
* Put nodeId,nodeoffset in pagemetadata for records
* Store continous node together
* Split page when node is added more than nodecapacity of page
* Put Page on different batchlevel when exceeding pagesize
* SequentializeFile : After every few days, read entire file and sequentialize node properly

func (node *BTreeNode) Search(key int) (bool, uint64) {
	node.mu.RLock()
	version := atomic.LoadUint64(&node.Version) // ✅ Read version first
	node.mu.RUnlock()

	// Search without locking
	for _, k := range node.Keys {
		if k == key {
			return true, version
		}
	}

	return false, version
}

func (node *BTreeNode) Insert(key int, oldVersion uint64) bool {
	node.mu.Lock()
	defer node.mu.Unlock()

	// ✅ Check if the version has changed
	if node.Version != oldVersion {
		return false // ❌ Conflict, retry needed
	}

	// ✅ Insert key
	node.Keys = append(node.Keys, key)
	sort.Ints(node.Keys)

	// ✅ Mark node as modified
	atomic.AddUint64(&node.Version, 1)

	return true
}

func (node *BTreeNode) InsertCAS(key int) bool {
	oldVersion := atomic.LoadUint64(&node.Version)

	// Modify the node
	node.mu.Lock()
	node.Keys = append(node.Keys, key)
	sort.Ints(node.Keys)
	node.mu.Unlock()

	// ✅ Use CAS to update the version only if it hasn't changed
	return atomic.CompareAndSwapUint64(&node.Version, oldVersion, oldVersion+1)
}

func (tree *BTree) SplitNode(node *BTreeNode) {
	node.mu.Lock()  // 🚦 Lock node before modifying
	defer node.mu.Unlock()

	mid := len(node.Keys) / 2
	newNode := &BTreeNode{
		Keys:   node.Keys[mid:],  // Move half the keys
		IsLeaf: node.IsLeaf,
	}
	node.Keys = node.Keys[:mid] // Retain first half

	// 🚦 Lock parent before modifying children
	tree.Root.mu.Lock()
	tree.Root.Children = append(tree.Root.Children, newNode)
	tree.Root.mu.Unlock()
}

func (node *BTreeNode) Delete(key int) bool {
	node.mu.Lock()
	defer node.mu.Unlock()

	for i, k := range node.Keys {
		if k == key {
			node.Keys = append(node.Keys[:i], node.Keys[i+1:]...) // Remove key
			return true
		}
	}
	return false
}

// ✅ Merge two nodes safely (Lock both)
func MergeNodes(left *BTreeNode, right *BTreeNode) {
	left.mu.Lock()
	right.mu.Lock()
	defer left.mu.Unlock()
	defer right.mu.Unlock()

	left.Keys = append(left.Keys, right.Keys...)
	right.Keys = nil // Mark as deleted
}

allChildren := make([]int32, numNodes*childrenSize) // One big allocation

for i := range nodes {
    nodes[i].children = allChildren[i*childrenSize : (i+1)*childrenSize] // No new allocation
}
```
* Delete key, if deletes node, keep deleted node in array for removal from disk
* Images, binary data visual
* Kademlia
* Persist to storage, with compression
* Bufferpool, Timebaseminheap
* Inverted tree, index, ngram, bm25
* hyperloglog, bloomfilter
* Wal
* Transaction concurrency
* zstd compress
