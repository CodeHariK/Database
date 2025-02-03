package secretary

import (
	"os"
	"sync"
)

const (
	SECRETARY                  = "SECRETARY"
	SECRETARY_HEADER_LENGTH    = 64
	MAX_COLLECTION_NAME_LENGTH = 30

	MIN_ORDER = uint8(3)   // Minimum allowed order for the B+ Tree
	MAX_ORDER = uint8(200) // Maximum allowed order for the B+ Tree

	BYTE_8  = 1<<8 - 1
	BYTE_16 = 1<<16 - 1

	RECORD_BATCH_OFFSET_AND = 1<<56 - 1
	RECORD_BATCH_LEVEL_AND  = BYTE_8 << 56

	NODE_BATCH_OFFSET_AND = 1<<48 - 1
	NODE_INDEX_AND        = (BYTE_16 - 1) << 48
)

type Secretary struct {
	tree map[string]*bTree
}

/*
**HEADER AND NODES**

------64 bytes------
SECRETARY				(9 bytes) 9
order 					(uint8)   10
keySize        			(uint8)   11
batchNumLevel  			(uint8)   12
batchBaseSize  			(uint32)  16
batchIncrement 			(uint8)   17
batchLength    			(uint8)   18
collectionName			(string)
---------------------
1  			Root
order		Internal
order ^ 2	Internal
order ^ 3	Internal
order ^ 4	Internal
...
order ^ n	Leaf
---------------------
*/
type bTree struct {
	CollectionName string `bin:"collectionName" max:"30"` // Max 30Char

	nodeBatchStore    *BatchStore
	recordBatchStores []*BatchStore

	root *Node // Root node of the tree

	Order          uint8  `bin:"order"`          // Max = 255, Order of the tree (maximum number of children)
	KeySize        uint8  `bin:"keySize"`        // 8 or 16 bytes
	BatchNumLevel  uint8  `bin:"batchNumLevel"`  // 32, Max 256 levels
	BatchBaseSize  uint32 `bin:"batchBaseSize"`  // 1024B
	BatchIncrement uint8  `bin:"batchIncrement"` // 125 => 1.25
	BatchLength    uint8  `bin:"batchLength"`    // 64 (2432*64/1024 = 152 KB), 128 (304KB), 431 (1 MB)
}

type BatchStore struct {
	file *os.File

	headerSize uint8
	level      uint8 // (1.25 ^ 0)MB  (1.25 ^ 1)MB  ... (1.25 ^ 31)MB

	batchSize uint32 // Maximum batch size = 4GB

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
