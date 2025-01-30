package secretary

import (
	"encoding/binary"
	"os"
)

func NewRecordManager(filename string) *RecordManager {
	file, _ := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0o644)
	stat, _ := file.Stat()
	return &RecordManager{
		file:       file,
		usedBlocks: stat.Size() / RECORD_BLOCK_SIZE,
	}
}

// Allocate new block with proper tracking
func (rm *RecordManager) allocateBlock(nodeID int64) int64 {
	rm.blockMutex.Lock()
	defer rm.blockMutex.Unlock()

	var blockID int64
	if len(rm.freeBlocks) > 0 {
		blockID = rm.freeBlocks[0]
		rm.freeBlocks = rm.freeBlocks[1:]
	} else {
		blockID = rm.usedBlocks
		rm.usedBlocks++
	}

	// Initialize block header
	header := BlockHeader{
		NodeID:    nodeID,
		NextBlock: -1,
		FreeSpace: RECORD_BLOCK_SIZE - HEADER_SIZE,
	}
	rm.writeHeader(blockID, header)
	return blockID
}

func (rm *RecordManager) StoreRecords(node *Node, records [][]byte) error {
	rm.nodeMutex.Lock()
	defer rm.nodeMutex.Unlock()

	currentBlockID := node.BlockIDs[len(node.BlockIDs)-1] // Try last block first
	var position int64 = HEADER_SIZE

	// Try existing blocks first
	if len(node.BlockIDs) > 0 {
		header := rm.readHeader(currentBlockID)
		position = RECORD_BLOCK_SIZE - int64(header.FreeSpace)
	}

	for _, record := range records {
		recordSize := 4 + len(record) // 4-byte length prefix

		// Check if record fits in current block
		if position+int64(recordSize) > RECORD_BLOCK_SIZE {
			// Allocate new block
			newBlockID := rm.allocateBlock(node.ID)

			// Link blocks
			header := rm.readHeader(currentBlockID)
			header.NextBlock = newBlockID
			rm.writeHeader(currentBlockID, header)

			currentBlockID = newBlockID
			node.BlockIDs = append(node.BlockIDs, newBlockID)
			position = HEADER_SIZE
		}

		// Write record
		rm.writeRecord(currentBlockID, position, record)
		position += int64(recordSize)

		// Update block metadata
		header := rm.readHeader(currentBlockID)
		header.RecordCount++
		header.FreeSpace -= uint32(recordSize)
		rm.writeHeader(currentBlockID, header)
	}

	return nil
}

// Get all blocks for a node with optional caching
func (rm *RecordManager) GetNodeBlocks(node *Node) [][]byte {
	rm.nodeMutex.RLock()
	defer rm.nodeMutex.RUnlock()

	var records [][]byte
	for _, blockID := range node.BlockIDs {
		header := rm.readHeader(blockID)
		position := HEADER_SIZE

		for i := 0; i < int(header.RecordCount); i++ {
			sizeBuf := make([]byte, 4)
			rm.file.ReadAt(sizeBuf, blockID*RECORD_BLOCK_SIZE+int64(position))
			recordSize := int(binary.LittleEndian.Uint32(sizeBuf))

			data := make([]byte, recordSize)
			rm.file.ReadAt(data, blockID*RECORD_BLOCK_SIZE+int64(position)+4)
			records = append(records, data)

			position += 4 + int64(recordSize)
		}
	}
	return records
}

// Free all blocks associated with a node
func (rm *RecordManager) FreeNodeBlocks(node *Node) {
	rm.blockMutex.Lock()
	defer rm.blockMutex.Unlock()

	rm.freeBlocks = append(rm.freeBlocks, node.BlockIDs...)
	node.BlockIDs = nil
}

// Helper methods
func (rm *RecordManager) writeHeader(blockID int64, header BlockHeader) {
	// Calculate checksum (simple example)
	header.Checksum = crc32(header)
	buf := make([]byte, HEADER_SIZE)
	binary.LittleEndian.PutUint64(buf[0:8], uint64(header.NodeID))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(header.NextBlock))
	binary.LittleEndian.PutUint32(buf[16:20], header.RecordCount)
	binary.LittleEndian.PutUint32(buf[20:24], header.FreeSpace)
	binary.LittleEndian.PutUint32(buf[24:28], header.Checksum)
	rm.file.WriteAt(buf, blockID*RECORD_BLOCK_SIZE)
}

func (rm *RecordManager) readHeader(blockID int64) BlockHeader {
	buf := make([]byte, HEADER_SIZE)
	rm.file.ReadAt(buf, blockID*RECORD_BLOCK_SIZE)
	return BlockHeader{
		NodeID:      int64(binary.LittleEndian.Uint64(buf[0:8])),
		NextBlock:   int64(binary.LittleEndian.Uint64(buf[8:16])),
		RecordCount: binary.LittleEndian.Uint32(buf[16:20]),
		FreeSpace:   binary.LittleEndian.Uint32(buf[20:24]),
		Checksum:    binary.LittleEndian.Uint32(buf[24:28]),
	}
}

func (rm *RecordManager) writeRecord(blockID int64, position int64, data []byte) {
	length := len(data)
	lengthBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBuf, uint32(length))

	offset := blockID*RECORD_BLOCK_SIZE + position
	rm.file.WriteAt(lengthBuf, offset)
	rm.file.WriteAt(data, offset+4)
}

// Simple checksum example
func crc32(header BlockHeader) uint32 {
	// Implement real CRC32 calculation in production
	return 0 // Placeholder
}
