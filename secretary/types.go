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

type NodeGroupBlock struct {
	/*
		NodeGroupID (4)  |  RecordCount (4)  |  UsedSpace (4)
		Records
	*/
	nodeGroup   NodeGroup
	RecordCount uint32 // 4 bytes
	UsedSpace   uint32 // 4 bytes

	Records *[]Record
}

type NodeGroup struct {
	id         uint32
	blockLevel int8 // 8 blockSize level
	leafIDs    []uint32
}

type InternalNode struct {
	parent *InternalNode // Parent node pointer
	next   *InternalNode // Pointer to next leaf node (for leaf nodes)
	prev   *InternalNode // Pointer to next leaf node (for leaf nodes)

	children []*InternalNode // Child pointers (for internal nodes)
}

type LeafNode struct {
	isLeaf      bool
	nodeGroupId uint32

	//
	parent *InternalNode // Parent node pointer
	next   *InternalNode // Pointer to next leaf node (for leaf nodes)
	prev   *InternalNode // Pointer to next leaf node (for leaf nodes)

	records *[]Record // Cached records
}

type Record struct {
	NodeId  uint32
	BlockID uint32
	Offset  uint32
	Size    uint32
	Key     uint64
	Value   []byte
}
