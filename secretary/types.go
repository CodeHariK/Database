package secretary

import (
	"os"
	"sync"
)

/*
-----------------------------------------------------------------------------
                              B+ Tree Structure (Order 4)
-----------------------------------------------------------------------------

After inserting [10, 20, 5, 6, 12, 30, 7, 17]:

                       [7, 10]
                      /   |   \
              [5,6,7]  [10,12]  [17,20,30]
               (Leaf)    (Leaf)     (Leaf)
                 |---------|---------→

Key Features:
• Internal nodes contain separator keys
• Leaf nodes contain actual keys
• Leaf nodes linked via `next` pointers
• Minimum keys per node = ⌈order/2⌉ - 1 = 1
• Maximum keys per node = order - 1 = 3

-----------------------------------------------------------------------------
                        Insertion/Deletion Visualization
-----------------------------------------------------------------------------

1. Initial Insertion (10, 20, 5, 6):

           [10]               → Split when inserting 6:
           /   \
  [5,6,10,20]             → Becomes:
         ↓
       [10]
      /     \
 [5,6]    [10,20]

2. After deleting 6 and 12:

          [10]
         /    \
    [5,7]  [17,20,30]

Leaf chain: [5,7] → [17,20,30]

-----------------------------------------------------------------------------
                          Range Query (5-17) Execution
-----------------------------------------------------------------------------

         [7,10]
        /   |   \
[5,6,7]→[10,12]→[17,20,30]

FindRange(5, 17):
1. Start at leftmost leaf [5,6,7]
2. Traverse via `next` pointers:
   - [5,6,7] → 5,6,7
   - [10,12] → 10,12
   - [17,20,30] → 17 (stop at 17)
Result: [5,6,7,10,12,17]

-----------------------------------------------------------------------------
                          Node Structure Details
-----------------------------------------------------------------------------

Internal Node:
+---------------+
| keys: [7,10]  |
| children: [*]  | → Points to child nodes
| leaf: false    |
| numKeys: 2     |
+---------------+

Leaf Node:
+-----------------------+
| keys: [5,6,7]         |
| next: → [10,12] node  |
| leaf: true            |
| numKeys: 3            |
+-----------------------+

-----------------------------------------------------------------------------
                          Operation Complexities
-----------------------------------------------------------------------------
Operation      | Time Complexity | Visual Representation
---------------+-----------------+----------------------
Insert         | O(log_t n)      | Root→...→Leaf path
Delete         | O(log_t n)      | Leaf→Parent rebalance
Search         | O(log_t n)      | Vertical traversal
Range Query    | O(log_t n + k)  | Horizontal leaf scan

-----------------------------------------------------------------------------
                          Legend
-----------------------------------------------------------------------------
[ ]     : Node
→       : Pointer
|       : Parent-child relationship
- - - - : Leaf node linkage
t       : Tree order (minimum degree)
k       : Number of keys in range

*/

const (
	RECORD_BLOCK_SIZE = 1 << 20 // 1MB
	HEADER_SIZE       = 28      // 28 bytes per block header

	MIN_ORDER     = 3   // Minimum allowed order for the B+ Tree
	MAX_ORDER     = 255 // Maximum allowed order for the B+ Tree
	DEFAULT_ORDER = 4   // Default order used if none is specified

	CHUNK_SIZE         = 1 << 12 // 4KB per node
	FILE_NODE_INTERNAL = "internal.bin"
	FILE_NODE_LEAF     = "leaf.bin"
	FILE_RECORDS       = "records.bin"
)

// Node metadata offsets (common)
const (
	IsLeafOffset   = 0
	NumKeysOffset  = 1
	ParentIDOffset = 5
)

// Internal node specific offsets
const (
	InternalKeysStart  = 13
	InternalChildStart = 13 + (MaxOrder-1)*4
	InternalNodeSize   = 13 + (MaxOrder-1)*4 + MaxOrder*8
)

// Leaf node specific offsets
const (
	LeafNextIDOffset = 13
	LeafKeysStart    = 21
	LeafRecordsStart = 21 + (MaxOrder-1)*4
	LeafNodeSize     = 21 + (MaxOrder-1)*4 + (MaxOrder-1)*8
)

type (
	Key int64
)

type Record struct {
	BlockID int32
	Offset  int32
	Size    int32
	Value   []byte
}

// Node represents a single node in the B+ Tree
type Node struct {
	/*
		Internal Node Chunk (8192 bytes)
		+----------------+----------------+----------------+----------------+
		| Metadata       | Keys           | Children       | Padding        |
		| (13 bytes)     | (MaxOrder-1)*4 | (MaxOrder*8)   |                |
		+----------------+----------------+----------------+----------------+

		Leaf Node Chunk (8192 bytes)
		+----------------+----------------+----------------+----------------+
		| Metadata       | Keys           | Record IDs     | Padding        |
		| (21 bytes)     | (MaxOrder-1)*4 | (MaxOrder-1)*8 |                |
		+----------------+----------------+----------------+----------------+
	*/
	id     int64
	isLeaf bool // Flag indicating if this is a leaf node

	blockLevel int8 // 8 blockSize level

	//
	parent   *Node   // Parent node pointer
	next     *Node   // Pointer to next leaf node (for leaf nodes)
	children []*Node // Child pointers (for internal nodes)

	//
	numKeys uint8    // Number of keys in the node
	keys    []Key    // For B+ tree operations
	records []Record // Cached records
}

// BPlusTree represents the B+ Tree structure
type BPlusTree struct {
	root  *Node // Root node of the tree
	order int   // Order of the tree (maximum number of children)

	internalFile *os.File
	leafFile     *os.File
	recordFile   *os.File

	averageRecordSize int32 // MaxRecordSize = 4GB
	numRecords        int64 // Total records in the tree

	freeInternals []int64
	freeLeaves    []int64
}

type BlockHeader struct {
	/*
		+----------------+----------------+----------------+----------------+----------------+
		| NodeID (8)     | NextBlock (8)  | RecordCount (4)| FreeSpace (4)  | Checksum (4)   |
		+----------------+----------------+----------------+----------------+----------------+
		| Records...                                                                         |
		+------------------------------------------------------------------------------------+
	*/
	NodeID      int64  // 8 bytes
	NextBlock   int64  // 8 bytes
	RecordCount uint32 // 4 bytes
	FreeSpace   uint32 // 4 bytes
	Checksum    uint32 // 4 bytes
}

type RecordManager struct {
	file       *os.File
	freeBlocks []int64      // Recycled block IDs
	usedBlocks int64        // Total blocks ever allocated
	blockMutex sync.Mutex   // Protects block allocation
	nodeMutex  sync.RWMutex // Protects node block lists
}
