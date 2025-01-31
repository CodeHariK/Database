package secretary

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

const (
	BlockSize  = 1024 // 1KB blocks
	HeaderSize = 4    // Free list entry (4 bytes per block index)
)

type BlockStorage struct {
	file     *os.File
	freeList []int64 // List of free blocks
	mu       sync.Mutex
}

// Open or create a file for block storage
func OpenStorage(filePath string) (*BlockStorage, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		return nil, err
	}

	storage := &BlockStorage{
		file:     file,
		freeList: []int64{},
	}

	// Scan file for free blocks
	storage.scanFreeList()

	return storage, nil
}

// Scan file for free blocks (find empty blocks)
func (bs *BlockStorage) scanFreeList() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	info, err := bs.file.Stat()
	if err != nil {
		fmt.Println("Error scanning file:", err)
		return
	}

	totalBlocks := info.Size() / BlockSize
	buffer := make([]byte, BlockSize)

	for i := int64(0); i < totalBlocks; i++ {
		offset := i * BlockSize
		_, err := bs.file.ReadAt(buffer, offset)
		if err == nil && isBlockEmpty(buffer) {
			bs.freeList = append(bs.freeList, i)
		}
	}
}

// Check if a block is empty (all zeroes)
func isBlockEmpty(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// Allocate a new block (reuse free block if available)
func (bs *BlockStorage) allocateBlock() int64 {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if len(bs.freeList) > 0 {
		block := bs.freeList[0]
		bs.freeList = bs.freeList[1:]
		return block
	}

	info, _ := bs.file.Stat()
	return info.Size() / BlockSize
}

// Write data to a specific block
func (bs *BlockStorage) Write(blockIndex int64, data []byte) error {
	if len(data) > BlockSize {
		return fmt.Errorf("data too large, must be at most %d bytes", BlockSize)
	}

	bs.mu.Lock()
	defer bs.mu.Unlock()

	offset := blockIndex * BlockSize
	buffer := make([]byte, BlockSize)
	copy(buffer, data)

	_, err := bs.file.WriteAt(buffer, offset)
	return err
}

// Read data from a specific block
func (bs *BlockStorage) Read(blockIndex int64) ([]byte, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	offset := blockIndex * BlockSize
	buffer := make([]byte, BlockSize)

	_, err := bs.file.ReadAt(buffer, offset)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// Delete a block by clearing its data and adding it to the free list
func (bs *BlockStorage) Delete(blockIndex int64) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	offset := blockIndex * BlockSize
	emptyBlock := make([]byte, BlockSize)

	_, err := bs.file.WriteAt(emptyBlock, offset)
	if err == nil {
		bs.freeList = append(bs.freeList, blockIndex)
	}

	return err
}

// Compact the file by removing empty blocks
func (bs *BlockStorage) Compact() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	info, err := bs.file.Stat()
	if err != nil {
		return err
	}

	totalBlocks := info.Size() / BlockSize
	newFile, err := os.CreateTemp("", "compact_storage")
	if err != nil {
		return err
	}
	defer newFile.Close()

	buffer := make([]byte, BlockSize)
	newOffset := int64(0)

	for i := int64(0); i < totalBlocks; i++ {
		offset := i * BlockSize
		_, err := bs.file.ReadAt(buffer, offset)
		if err == nil && !isBlockEmpty(buffer) {
			_, err = newFile.WriteAt(buffer, newOffset)
			if err != nil {
				return err
			}
			newOffset += BlockSize
		}
	}

	bs.file.Close()
	err = os.Rename(newFile.Name(), bs.file.Name())
	if err != nil {
		return err
	}

	bs.file, err = os.OpenFile(bs.file.Name(), os.O_RDWR, 0o666)
	bs.freeList = []int64{} // Reset free list
	bs.scanFreeList()

	return err
}

// Close the file
func (bs *BlockStorage) Close() error {
	return bs.file.Close()
}

// Store an integer as bytes
func intToBytes(n int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(n))
	return buf
}

// Retrieve an integer from bytes
func bytesToInt(buf []byte) int64 {
	return int64(binary.LittleEndian.Uint64(buf))
}

func main() {
	// Open or create storage file
	storage, err := OpenStorage("storage.dat")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer storage.Close()

	// Allocate and write to a new block
	block := storage.allocateBlock()
	data := []byte("Hello, block storage!")
	err = storage.Write(block, data)
	if err != nil {
		fmt.Println("Write Error:", err)
		return
	}

	// Read from the block
	readData, err := storage.Read(block)
	if err != nil {
		fmt.Println("Read Error:", err)
		return
	}
	fmt.Printf("Read Data: %s\n", readData)

	// Delete the block
	err = storage.Delete(block)
	if err != nil {
		fmt.Println("Delete Error:", err)
		return
	}

	// Compact storage
	err = storage.Compact()
	if err != nil {
		fmt.Println("Compact Error:", err)
		return
	}

	fmt.Println("Storage operations completed successfully!")
}
