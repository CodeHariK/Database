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
batchNumLevel  			(uint8)    11
batchBaseSize  			(uint32)   15
batchIncrement 			(uint8)    16
batchLength    			(uint8)    17
nodeSeq    				(uint64)   25
keySeq    				(uint64)   33
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
	CollectionName string `json:"collectionName" bin:"collectionName" max:"30"` // Max 30Char

	indexPager   *Pager
	recordPagers []*Pager

	root *Node // Root node of the tree

	Order          uint8  `json:"order" bin:"order"`                   // Max = 255, Order of the tree (maximum number of children)
	BatchNumLevel  uint8  `json:"batchNumLevel" bin:"batchNumLevel"`   // 32, Max 256 levels
	BatchBaseSize  uint32 `json:"batchBaseSize" bin:"batchBaseSize"`   // 1024MB
	BatchIncrement uint8  `json:"batchIncrement" bin:"batchIncrement"` // 125 => 1.25
	BatchLength    uint8  `json:"batchLength" bin:"batchLength"`       // 64 (2432*64/1024 = 152 KB), 128 (304KB), 431 (1 MB)
	NodeSeq        uint64 `json:"nodeSeq" bin:"nodeSeq"`               // Incrementing Node sequence
	KeySeq         uint64 `json:"keySeq" bin:"keySeq"`                 // Incrementing Key sequence

	nodeSize   uint32
	minNumKeys uint32 // Minimum required keys in node
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
	NodeID uint64 `bin:"NodeID"` // Unique ID for the node

	parent   *Node
	next     *Node
	prev     *Node
	children []*Node
	records  []*Record

	Offset       DataLocation
	ParentOffset DataLocation `bin:"ParentOffset"`
	NextOffset   DataLocation `bin:"NextOffset"`
	PrevOffset   DataLocation `bin:"PrevOffset"`

	// NumKeys    uint8          `bin:"NumKeys"`
	KeyOffsets []DataLocation `bin:"KeyOffsets"`               // (8 bytes) [node children offset | record offset]
	Keys       [][]byte       `bin:"Keys" array_elem_len:"16"` // (16 bytes)
}

// Page represents a single fixed-size page in memory.
type Page struct {
	Index int64
	Data  []byte

	mu sync.Mutex // Per-page lock
}

// Pager manages reading and writing pages.
type Pager struct {
	file *os.File

	headerSize int
	level      uint8 // (1.25 ^ 0)MB  (1.25 ^ 1)MB  ... (1.25 ^ 31)MB

	pageSize int64 // Maximum batch size = 4GB

	cache      *ristretto.Cache[int64, *Page] // In-memory cache
	dirtyPages map[int64]bool

	mu sync.Mutex
}

// Record
// 56 bit Offset
// 8  bit BatchLevel
//
// Node
// 48 bit BatchOffset (Max batch in file = 2^48)
// 16 bit NodeIndex	   (Max nodes in batch = 2^16)
type DataLocation int64

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
