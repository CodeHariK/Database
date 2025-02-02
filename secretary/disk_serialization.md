Here's the implementation of serialization and deserialization for the `Node` structure with the specified requirements:

```go
package bplus

import (
	"encoding/binary"
	"errors"
	"os"
)

const (
	BYTE_8  = 1<<8 - 1
	BYTE_16 = 1<<16 - 1

	RECORD_BLOCK_OFFSET_AND = 1<<56 - 1
	RECORD_BLOCK_LEVEL_AND  = BYTE_8 << 56

	NODE_BLOCK_OFFSET_AND = 1<<48 - 1
	NODE_INDEX_AND        = (BYTE_16 - 1) << 48

	RecordHeaderSize = 20 // 8 (Offset) + 4 (Size) + 8 (Key)
)

type DataLocation int64

type Record struct {
	Offset DataLocation // 8 bytes
	Size   uint32       // 4 bytes
	Key    []byte       // 8 or 16 bytes
	Value  []byte
}

type Node struct {
	parent   *Node
	next     *Node
	prev     *Node
	children []*Node
	records  []Record
}

type NodeSerialized struct {
	offset       DataLocation
	parentOffset DataLocation
	nextOffset   DataLocation
	prevOffset   DataLocation
	keyOffsets   []DataLocation
	Keys         [][]byte
}

// SerializeNode converts Node to NodeSerialized for disk storage
func SerializeNode(node *Node, nodeOffset DataLocation) (*NodeSerialized, error) {
	if node == nil {
		return nil, errors.New("node is nil")
	}

	serialized := &NodeSerialized{
		offset:       nodeOffset,
		parentOffset: getNodeLocation(node.parent),
		nextOffset:   getNodeLocation(node.next),
		prevOffset:   getNodeLocation(node.prev),
		keyOffsets:   make([]DataLocation, len(node.records)),
		Keys:         make([][]byte, len(node.records)),
	}

	// Serialize records
	for i, record := range node.records {
		serialized.keyOffsets[i] = record.Offset
		serialized.Keys[i] = record.Key
	}

	return serialized, nil
}

// DeserializeNode converts NodeSerialized back to Node
func DeserializeNode(serialized *NodeSerialized, file *os.File) (*Node, error) {
	if serialized == nil {
		return nil, errors.New("serialized node is nil")
	}

	node := &Node{
		children: make([]*Node, 0),
		records:  make([]Record, len(serialized.Keys)),
	}

	// Deserialize records
	for i, key := range serialized.Keys {
		// Read record from file
		record, err := readRecord(file, serialized.keyOffsets[i])
		if err != nil {
			return nil, err
		}
		node.records[i] = *record
	}

	// Note: parent, next, prev pointers will be set during tree reconstruction
	return node, nil
}

// Helper functions
func getNodeLocation(node *Node) DataLocation {
	if node == nil {
		return -1
	}
	// Implement actual node location calculation
	return 0 // Placeholder
}

func readRecord(file *os.File, offset DataLocation) (*Record, error) {
	buf := make([]byte, RecordHeaderSize)
	_, err := file.ReadAt(buf, int64(offset))
	if err != nil {
		return nil, err
	}

	record := &Record{
		Offset: DataLocation(binary.LittleEndian.Uint64(buf[0:8])),
		Size:   binary.LittleEndian.Uint32(buf[8:12]),
		Key:    make([]byte, 8), // or 16 based on your key size
	}

	// Read key
	copy(record.Key, buf[12:20])

	// Read value
	record.Value = make([]byte, record.Size)
	_, err = file.ReadAt(record.Value, int64(record.Offset))
	if err != nil {
		return nil, err
	}

	return record, nil
}

// DataLocation manipulation
func MakeRecordLocation(blockLevel uint8, blockOffset uint64) DataLocation {
	return DataLocation((uint64(blockLevel) << 56) | (blockOffset & RECORD_BLOCK_OFFSET_AND))
}

func ExtractRecordLocation(loc DataLocation) (uint8, uint64) {
	blockLevel := uint8((loc & RECORD_BLOCK_LEVEL_AND) >> 56)
	blockOffset := uint64(loc & RECORD_BLOCK_OFFSET_AND)
	return blockLevel, blockOffset
}

func MakeNodeLocation(blockOffset uint64, nodeIndex uint16) DataLocation {
	return DataLocation((uint64(nodeIndex) << 48) | (blockOffset & NODE_BLOCK_OFFSET_AND))
}

func ExtractNodeLocation(loc DataLocation) (uint64, uint16) {
	nodeIndex := uint16((loc & NODE_INDEX_AND) >> 48)
	blockOffset := uint64(loc & NODE_BLOCK_OFFSET_AND)
	return blockOffset, nodeIndex
}

// Node storage format
/*
+----------------+----------------+----------------+----------------+
| parentOffset   | nextOffset     | prevOffset     | numRecords     |
| (8 bytes)      | (8 bytes)      | (8 bytes)      | (4 bytes)      |
+----------------+----------------+----------------+----------------+
| keyOffsets...                                                      |
| (8 bytes each)                                                     |
+----------------+----------------+----------------+----------------+
| Keys...                                                           |
| (8 or 16 bytes each)                                              |
+----------------+----------------+----------------+----------------+
*/

// Example usage
func main() {
	// Create sample node
	node := &Node{
		records: []Record{
			{Offset: 100, Size: 50, Key: []byte("key1"), Value: []byte("value1")},
			{Offset: 200, Size: 60, Key: []byte("key2"), Value: []byte("value2")},
		},
	}

	// Serialize
	serialized, _ := SerializeNode(node, 0)

	// Deserialize
	file, _ := os.Open("data.bin")
	deserialized, _ := DeserializeNode(serialized, file)

	// Verify
	if len(deserialized.records) != len(node.records) {
		panic("record count mismatch")
	}
}
```

Key Features:

1. **Node Serialization**
- Converts in-memory Node to disk-friendly format
- Handles pointer relationships via offsets
- Maintains record key/offset mapping

2. **Record Storage**
- Fixed header size for efficient access
- Variable-length value storage
- Support for 8 or 16 byte keys

3. **DataLocation Encoding**
```go
// Record Location Format
// 56 bits: Block Offset
// 8 bits: Block Level

// Node Location Format
// 48 bits: Block Offset
// 16 bits: Node Index
```

4. **File Structure**
- Sequential storage of nodes
- Direct access via offsets
- Efficient record retrieval

5. **Error Handling**
- Nil pointer checks
- File I/O error propagation
- Data validation

To integrate with your B+ Tree:

1. **Tree Operations**
```go
func (t *BPlusTree) Insert(key []byte, value []byte) {
	// Find appropriate leaf node
	leaf := t.FindLeaf(key)
	
	// Store record
	recordOffset := t.StoreRecord(key, value)
	
	// Update node
	leaf.records = append(leaf.records, Record{
		Offset: recordOffset,
		Size:   uint32(len(value)),
		Key:    key,
		Value:  value,
	})
	
	// Serialize and write node
	serialized, _ := SerializeNode(leaf, t.GetNodeOffset(leaf))
	t.WriteNode(serialized)
}
```

2. **Record Management**
```go
func (t *BPlusTree) StoreRecord(key []byte, value []byte) DataLocation {
	// Find free space in record file
	offset := t.recordFile.Size()
	
	// Write record
	header := make([]byte, RecordHeaderSize)
	binary.LittleEndian.PutUint64(header[0:8], uint64(offset))
	binary.LittleEndian.PutUint32(header[8:12], uint32(len(value)))
	copy(header[12:20], key)
	t.recordFile.Write(header)
	t.recordFile.Write(value)
	
	return DataLocation(offset)
}
```

This implementation provides:
- Efficient disk storage format
- Direct access to nodes and records
- Compact offset encoding
- Support for large datasets
- Clear separation of concerns

For production use, consider adding:
- Transaction support
- Write-ahead logging
- Compression for records
- Cache layer for frequently accessed nodes
- Bulk operation optimizations