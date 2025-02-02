package secretary

import (
	"os"
	"sync"
)

type Secretary struct {
	tree map[string]*BPlusTree
}

type BPlusTree struct {
	collectionName string

	nodeBatchStore    BatchStore
	recordBatchStores []BatchStore

	root *Node // Root node of the tree

	order          uint8  // Max = 255, Order of the tree (maximum number of children)
	keySize        uint8  // 8 or 16 bytes
	batchNumLevel  uint8  // 32, Max 256 levels
	batchBaseSize  uint32 // 1024B
	batchIncrement uint8  // 125 => 1.25
	batchLength    uint8  // 64 (2432*64/1024 = 152 KB), 128 (304KB), 431 (1 MB)
}

type BatchStore struct {
	file *os.File

	level uint8 // (1.25 ^ 0)MB  (1.25 ^ 1)MB  ... (1.25 ^ 31)MB

	size uint32 // Maximum batch size = 4GB

	mu sync.Mutex
}

/*
**Node Structure**
+----------------+----------------+----------------+----------------+
| parentOffset   | nextOffset     | prevOffset     | numKeys        |
| (8 bytes)      | (8 bytes)      | (8 bytes)      | (1 bytes)      |
+----------------+----------------+----------------+----------------+
| keyOffsets...                                                     |
| (8 bytes each)                                                    |
+----------------+----------------+----------------+----------------+
| Keys...                                                           |
| (8 or 16 bytes each)                                              |
+----------------+----------------+----------------+----------------+
*/
type Node struct {
	parent   *Node
	next     *Node
	prev     *Node
	children []*Node
	records  []*Record

	offset       DataLocation
	parentOffset DataLocation
	nextOffset   DataLocation
	prevOffset   DataLocation

	numKeys    uint8
	keyOffsets []DataLocation // (8 bytes)
	keys       [][]byte       // (8 bytes or 16 bytes)
}

func (n *Node) IsLeaf() bool {
	return len(n.children) == 0
}

const (
	BYTE_8  = 1<<8 - 1
	BYTE_16 = 1<<16 - 1

	RECORD_BATCH_OFFSET_AND = 1<<56 - 1
	RECORD_BATCH_LEVEL_AND  = BYTE_8 << 56

	NODE_BATCH_OFFSET_AND = 1<<48 - 1
	NODE_INDEX_AND        = (BYTE_16 - 1) << 48
)

// Record
// 56 bit Offset
// 8  bit BatchLevel
//
// Node
// 48 bit BatchOffset (Max batch in file = 2^48)
// 16 bit NodeIndex	   (Max nodes in batch = 2^16)
type DataLocation uint64

type Record struct {
	Offset DataLocation // (8 bytes)
	Size   uint32       // (4 bytes) Max size = 4GB
	Key    []byte       // (8 bytes or 16 bytes)
	Value  []byte
}

/**
**Node File Layout**
```
+----------------+----------------+----------------+----------------+
| Root Offset    | Node 1         | Node 2         | ...            |
| (8 bytes)      | (Variable)     | (Variable)     |                |
+----------------+----------------+----------------+----------------+
```

**Record File Layout**
```
+----------------+----------------+----------------+----------------+
| Record 1       | Record 2       | Record 3       | ...            |
| (Variable)     | (Variable)     | (Variable)     |                |
+----------------+----------------+----------------+----------------+
```

**Record Structure**
```
+----------------+----------------+----------------+----------------+
| Offset         | Size           | Key            | Value          |
| (8 bytes)      | (4 bytes)      | (8/16 bytes)   | (Variable)     |
+----------------+----------------+----------------+----------------+
```
*/
