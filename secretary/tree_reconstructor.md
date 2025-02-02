To reconstruct the B+ Tree from the serialized nodes on disk, we need to implement a tree reconstruction process. This involves:

1. **Reading Nodes**: Load serialized nodes from disk
2. **Rebuilding Relationships**: Reconstruct parent/child/next/prev pointers
3. **Reconstructing Records**: Load associated records
4. **Rebuilding Tree Structure**: Recreate the hierarchical structure

Here's the implementation:

```go
package bplus

import (
	"encoding/binary"
	"os"
)

type TreeReconstructor struct {
	file        *os.File
	recordFile  *os.File
	nodeMap     map[DataLocation]*Node
	rootOffset  DataLocation
}

func NewTreeReconstructor(nodeFile, recordFile string) *TreeReconstructor {
	nf, _ := os.Open(nodeFile)
	rf, _ := os.Open(recordFile)
	return &TreeReconstructor{
		file:       nf,
		recordFile: rf,
		nodeMap:    make(map[DataLocation]*Node),
	}
}

func (tr *TreeReconstructor) ReconstructTree() (*Node, error) {
	// Read root node offset from file header
	rootOffset := tr.readRootOffset()
	tr.rootOffset = rootOffset

	// Reconstruct all nodes
	tr.reconstructNodes()

	// Rebuild tree structure
	return tr.rebuildTreeStructure()
}

func (tr *TreeReconstructor) reconstructNodes() {
	// Start from root and traverse all nodes
	queue := []DataLocation{tr.rootOffset}
	for len(queue) > 0 {
		currentOffset := queue[0]
		queue = queue[1:]

		// Skip already processed nodes
		if _, exists := tr.nodeMap[currentOffset]; exists {
			continue
		}

		// Read and deserialize node
		node := tr.readNode(currentOffset)
		tr.nodeMap[currentOffset] = node

		// Add children to processing queue
		if !node.IsLeaf() {
			for _, childOffset := range node.childOffsets {
				if childOffset != -1 {
					queue = append(queue, childOffset)
				}
			}
		}
	}
}

func (tr *TreeReconstructor) rebuildTreeStructure() (*Node, error) {
	// Reconstruct parent/child relationships
	for offset, node := range tr.nodeMap {
		// Reconstruct parent
		if node.parentOffset != -1 {
			node.parent = tr.nodeMap[node.parentOffset]
		}

		// Reconstruct children
		if !node.IsLeaf() {
			node.children = make([]*Node, len(node.childOffsets))
			for i, childOffset := range node.childOffsets {
				if childOffset != -1 {
					node.children[i] = tr.nodeMap[childOffset]
				}
			}
		}

		// Reconstruct next/prev pointers
		if node.nextOffset != -1 {
			node.next = tr.nodeMap[node.nextOffset]
		}
		if node.prevOffset != -1 {
			node.prev = tr.nodeMap[node.prevOffset]
		}
	}

	return tr.nodeMap[tr.rootOffset], nil
}

func (tr *TreeReconstructor) readRootOffset() DataLocation {
	var rootOffset int64
	binary.Read(tr.file, binary.LittleEndian, &rootOffset)
	return DataLocation(rootOffset)
}

func (tr *TreeReconstructor) readNode(offset DataLocation) *Node {
	// Seek to node position
	tr.file.Seek(int64(offset), 0)

	// Read header
	header := make([]byte, NodeHeaderSize)
	tr.file.Read(header)

	node := &Node{
		offset:       offset,
		parentOffset: DataLocation(binary.LittleEndian.Uint64(header[0:8])),
		nextOffset:   DataLocation(binary.LittleEndian.Uint64(header[8:16])),
		prevOffset:   DataLocation(binary.LittleEndian.Uint64(header[16:24])),
		numRecords:   binary.LittleEndian.Uint32(header[24:28]),
	}

	// Read key offsets
	node.keyOffsets = make([]DataLocation, node.numRecords)
	for i := 0; i < int(node.numRecords); i++ {
		var offset int64
		binary.Read(tr.file, binary.LittleEndian, &offset)
		node.keyOffsets[i] = DataLocation(offset)
	}

	// Read keys
	node.Keys = make([][]byte, node.numRecords)
	for i := 0; i < int(node.numRecords); i++ {
		key := make([]byte, KeySize)
		tr.file.Read(key)
		node.Keys[i] = key
	}

	// Read child offsets for internal nodes
	if !node.IsLeaf() {
		node.childOffsets = make([]DataLocation, node.numRecords+1)
		for i := 0; i <= int(node.numRecords); i++ {
			var offset int64
			binary.Read(tr.file, binary.LittleEndian, &offset)
			node.childOffsets[i] = DataLocation(offset)
		}
	}

	// Load records
	node.records = make([]Record, node.numRecords)
	for i, keyOffset := range node.keyOffsets {
		record, _ := tr.readRecord(keyOffset)
		node.records[i] = *record
	}

	return node
}

func (tr *TreeReconstructor) readRecord(offset DataLocation) (*Record, error) {
	// Seek to record position
	tr.recordFile.Seek(int64(offset), 0)

	// Read record header
	header := make([]byte, RecordHeaderSize)
	tr.recordFile.Read(header)

	record := &Record{
		Offset: DataLocation(binary.LittleEndian.Uint64(header[0:8])),
		Size:   binary.LittleEndian.Uint32(header[8:12]),
		Key:    make([]byte, KeySize),
	}

	// Read key
	copy(record.Key, header[12:20])

	// Read value
	record.Value = make([]byte, record.Size)
	tr.recordFile.Read(record.Value)

	return record, nil
}

// Example usage
func main() {
	reconstructor := NewTreeReconstructor("nodes.bin", "records.bin")
	root, err := reconstructor.ReconstructTree()
	if err != nil {
		panic(err)
	}

	// Now you have the full reconstructed tree
	// with all relationships restored
}
```
