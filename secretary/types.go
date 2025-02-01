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

	blockStoreInternal BlockStore
	blockStoreLeaf     BlockStore
	blockStoreRecords  [32]BlockStore

	root *Node // Root node of the tree

	order                 uint8  // Max = 255, Order of the tree (maximum number of children)
	keySize               uint8  // 8 or 16 bytes
	blockSize             uint32 // ~1MB
	blockFactorPercentage uint8
	nodeGroupLength       uint32

	averageRecordSize uint32 // MaxRecordSize = 4GB
	RecordCount       uint64 // Total records in the tree
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

type Node struct {
	parent *Node
	next   *Node
	prev   *Node

	children []*Node

	records []Record
}

type NodeSerialised struct {
	offset       DataLocation
	parentOffset DataLocation
	nextOffset   DataLocation
	prevOffset   DataLocation

	keyOffsets []DataLocation // (8 bytes) 8 bits if Record for blockLevel, 48 bits for blockOffset
	Keys       [][]byte       // (8 bytes or 16 bytes)
}

// Record
// 48 bit Offset
// 16 bit BlockLevel
//
// Node
// 48 bit BlockID (Max blocks in file = 2^48)
// 16 bit NodeIndex if Node (Max nodes in block = 2^16)
type DataLocation int64

type Record struct {
	Offset DataLocation // (8 bytes)
	Size   uint32       // (4 bytes) Max size = 4GB
	Key    []byte       // (8 bytes or 16 bytes)
	Value  []byte
}
