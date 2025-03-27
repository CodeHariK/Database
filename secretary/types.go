package secretary

import (
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/dgraph-io/ristretto/v2"
)

const (
	SECRETARY                  = "SECRETARY"
	SECRETARY_HEADER_LENGTH    = 128
	MAX_COLLECTION_NAME_LENGTH = 30

	MIN_ORDER = 3   // Minimum allowed order for the B+ Tree
	MAX_ORDER = 200 // Maximum allowed order for the B+ Tree

	KEY_SIZE        = 16
	KEY_OFFSET_SIZE = 8
	POINTER_SIZE    = 8

	BYTE_8  = uint64(1<<8 - 1)
	BYTE_16 = uint64(1<<16 - 1)

	RECORD_BATCH_OFFSET_AND = uint64(1<<56 - 1)
	RECORD_BATCH_LEVEL_AND  = BYTE_8 << 56

	NODE_BATCH_OFFSET_AND = uint64(1<<48 - 1)
	NODE_INDEX_AND        = BYTE_16 << 48
)

type Secretary struct {
	trees map[string]*BTree

	listener net.Listener
	server   *http.Server

	httpClient http.Client

	quit chan any
	wg   sync.WaitGroup
	once sync.Once
}

/*
**HEADER AND NODES**

------128 bytes------
SECRETARY				(9 bytes)  9
order 					(uint8)    10
NumLevel  				(uint8)    11
baseSize  			(uint32)   15
increment 				(uint8)    16
nodeSeq    				(uint64)   24
numNodeSeq    			(uint64)   32
compactionBatchSize    	(uint32)   36
collectionName			(string)
---------------------
1  			Root		(5*1024 = 5120)
---------------------
order		Internal
order ^ 2	Internal
order ^ 3	Internal
order ^ 4	Internal
...
order ^ n	Leaf
---------------------
*/
type BTree struct {
	mu sync.Mutex // Global Lock for root changes

	CollectionName string `json:"collectionName" bin:"collectionName" max:"30"` // Max 30Char

	nodePager    *NodePager
	recordPagers []*RecordPager

	root               *Node // Root node of the tree
	nextCompactionNode *Node // Compaction Node For Current Batch

	Order     uint8  `json:"order" bin:"order"`         // Max = 255, Order of the tree (maximum number of children)
	NumLevel  uint8  `json:"numLevel" bin:"numLevel"`   // 32, Max 256 levels
	BaseSize  uint32 `json:"baseSize" bin:"baseSize"`   // 1024Bytes
	Increment uint8  `json:"increment" bin:"increment"` // 125 => 1.25

	NodeSeq    uint64 `json:"nodeSeq" bin:"nodeSeq"` // Incrementing Node sequence
	NumNodeSeq uint64 `json:"numNodeSeq" bin:"numNodeSeq"`

	nodeSize   uint32
	minNumKeys uint32 // Minimum required keys in node

	CompactionBatchSize uint32 `json:"compactionBatchSize" bin:"compactionBatchSize"`
}

/*
**Node Structure**
+----------------+----------------+----------------+----------------+
| NodeID         | parentOffset   | nextOffset     | prevOffset     |
| (8 bytes)      | (8 bytes)      | (8 bytes)      | (8 bytes)      |
+----------------+----------------+----------------+----------------+
| keyOffsets...                                                     |
| (8 bytes each)                                                    |
+----------------+----------------+----------------+----------------+
| Keys...                                                           |
| (16 bytes each)                                                   |
+----------------+----------------+----------------+----------------+
*/
type Node struct {
	mu      sync.RWMutex // 🚦 Latch for Synchronization
	Version uint64       `bin:"Version"` // ✅ OCC Version Number

	NodeID uint64 `bin:"NodeID"` // Unique ID for the node

	parent   *Node
	next     *Node
	prev     *Node
	children []*Node
	records  []*Record

	Index       uint64 `bin:"Index"`
	ParentIndex uint64 `bin:"ParentIndex"`
	NextIndex   uint64 `bin:"NextIndex"`
	PrevIndex   uint64 `bin:"PrevIndex"`

	KeyLocation []uint64 `bin:"KeyOffsets"`               // (8 bytes) [node children offset | record offset]
	Keys        [][]byte `bin:"Keys" array_elem_len:"16"` // (16 bytes)
}

type PageItem[T any] interface {
	NewPage(index int64) *Page[T]
	ToBytes() ([]byte, error)
	FromBytes([]byte) error
}

// Page represents a single fixed-size page in memory.
type Page[T any] struct {
	Index int64
	Data  T

	mu sync.Mutex // Per-page lock
}

// Pager manages reading and writing pages.
type Pager[T PageItem[T]] struct {
	file *os.File

	level uint8 // (1.25 ^ 0)MB  (1.25 ^ 1)MB  ... (1.25 ^ 31)MB

	headerSize int64
	itemSize   int64 // Maximum batch size = 4GB

	cache      *ristretto.Cache[int64, *Page[T]] // In-memory cache
	dirtyPages map[int64]bool

	mu sync.Mutex
}

type NodePager struct {
	*Pager[*Node]
}

type RecordPager struct {
	*Pager[*Record]
}

type Record struct {
	Offset uint64 // (8 bytes)
	Size   uint32 // (4 bytes) Max size = 4GB
	Key    []byte // (8 bytes or 16 bytes)
	Value  []byte
}

type RecordLocation struct {
	batchLevel uint8
	offset     uint64
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
