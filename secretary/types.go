package secretary

import (
	"os"
)

const (
	MIN_ORDER     = 3   // Minimum allowed order for the B+ Tree
	MAX_ORDER     = 255 // Maximum allowed order for the B+ Tree
	DEFAULT_ORDER = 4   // Default order used if none is specified

	FILE_NODE_INTERNAL = "internal.bin"
	FILE_NODE_LEAF     = "leaf.bin"
	FILE_RECORDS       = "records.bin"
)

type Secretary struct {
	tree map[string]*BPlusTree
}

// BPlusTree represents the B+ Tree structure
type BPlusTree struct {
	collectionName string

	root  *Node // Root node of the tree
	order int   // Order of the tree (maximum number of children)

	internalBlockStore NodeGroupBlockStore
	leafBlockStore     NodeGroupBlockStore
	recordsBlockStore  [32]NodeGroupBlockStore

	averageRecordSize int32 // MaxRecordSize = 4GB
	numRecords        int64 // Total records in the tree

	blockFactorPercentage int8
	nodeGroupLength       int8
}

type NodeGroupBlockStore struct {
	file *os.File

	blockLevel int8

	averageRecordSize int32 // MaxRecordSize = 4GB
	numRecords        int64 // Total records in the tree

	blockSize int32 // Maximum block size = 4GB
}

type NodeGroupBlock struct {
	/*
		NodeGroupID (4)  |  RecordCount (4)  |  UsedSpace (4)
		Records
	*/
	nodeGroup   NodeGroup
	RecordCount uint32 // 4 bytes
	UsedSpace   uint32 // 4 bytes

	Records []Record
}

type NodeGroup struct {
	ids []uint32
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
	id     uint32
	isLeaf bool // Flag indicating if this is a leaf node

	blockLevel int8 // 8 blockSize level

	//
	parent   *Node   // Parent node pointer
	next     *Node   // Pointer to next leaf node (for leaf nodes)
	prev     *Node   // Pointer to next leaf node (for leaf nodes)
	children []*Node // Child pointers (for internal nodes)

	//
	numKeys uint8    // Number of keys in the node
	records []Record // Cached records
}

type Record struct {
	NodeId  uint32
	BlockID uint32
	Offset  uint32
	Size    uint32
	Key     uint64
	Value   []byte
}
