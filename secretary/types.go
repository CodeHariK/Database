package secretary

import (
	"os"
	"sync"
)

type Secretary struct {
	tree map[string]*BPlusTree
}

// BPlusTree represents the B+ Tree structure
type BPlusTree struct {
	collectionName string

	root  *InternalNode // Root node of the tree
	order uint8         // Order of the tree (maximum number of children)

	blockStoreInternal BlockStore
	blockStoreLeaf     BlockStore
	blockStoreRecords  [32]BlockStore

	averageRecordSize uint32 // MaxRecordSize = 4GB
	RecordCount       uint64 // Total records in the tree

	blockFactorPercentage uint8
	nodeGroupLength       uint8
}

type BlockStore struct {
	file *os.File

	blockLevel int8

	averageRecordSize int32 // MaxRecordSize = 4GB
	RecordCount       int64 // Total records in the tree

	blockSize int32 // Maximum block size = 4GB

	freeList []int64

	mu sync.Mutex
}

type InternalNode struct {
	parent *InternalNode // Parent node pointer
	next   *InternalNode // Pointer to next leaf node (for leaf nodes)
	prev   *InternalNode // Pointer to next leaf node (for leaf nodes)

	children []*InternalNode // Child pointers (for internal nodes)

	records []Record
}

type InternalNodeSerialised struct {
	offset       uint64
	parentOffset uint64
	nextOffset   uint64
	prevOffset   uint64

	keyOffsets []KeyOffset // fixed "order" number of children
}

type KeyOffset struct {
	Offset uint64 // (8 bytes) 8 bits if Record for blockLevel, 48 bits for blockOffset
	Key    uint64 // (8 bytes)
}

type Record struct {
	Offset uint64 // (8 bytes) 8 bits if Record for blockLevel, 48 bits for blockOffset
	Key    uint64 // (8 bytes)
	Size   uint32 // (4 bytes) Max size = 4GB
	Value  []byte
}
