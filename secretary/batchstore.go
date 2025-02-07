package secretary

import (
	"fmt"
	"math"
	"os"
)

// Opens or creates a file and sets up the BatchStore
func (tree *BTree) NewBatchStore(fileType string, level uint8) (*BatchStore, error) {
	batchSize := uint32(float64(tree.BatchBaseSize) * math.Pow(float64(tree.BatchIncrement)/100, float64(level)))

	headerSize := 0

	path := fmt.Sprintf("SECRETARY/%s/%s_%d_%d.bin", tree.CollectionName, fileType, level, batchSize)
	if fileType == "index" {

		headerSize = SECRETARY_HEADER_LENGTH + int(tree.nodeSize)
		batchSize = 1024 * 1024

		path = fmt.Sprintf("SECRETARY/%s/%s.bin", tree.CollectionName, fileType)

	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		file.Close()
		// fmt.Printf("Create File : %s\n", path)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Open File : %s\n", path)

	return &BatchStore{
		file:       file,
		level:      level,
		headerSize: headerSize,
		batchSize:  batchSize,
	}, nil
}

// AllocateBatch writes zeroed data in chunks of pageSize for alignment
func (store *BatchStore) AllocateBatch(numBatch int32) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Get current file size
	fileInfo, err := store.file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	// Align to the next page boundary
	if (fileSize % int64(store.batchSize)) != 0 {
		return ErrorFileNotAligned(fileInfo)
	}

	// Expand file by writing zeros
	zeroBuf := make([]byte, store.batchSize*uint32(numBatch))
	_, err = store.file.WriteAt(zeroBuf, fileSize)
	if err != nil {
		return err
	}

	return nil
}

/**
* Header
* |
* |
* |
* |
* |
* |
*-----
*
*
*
*-----
*
* Offset
*          +
*-----     + Data shoud not exceed batch boundary
*          +
*
*
*-----
*
**/
// WriteAt writes data at the specified offset in the file.
// If there is not enough free space, it allocates a new batch.
func (store *BatchStore) WriteAt(offset int64, data []byte) error {
	// Ensure data size does not exceed batchSize
	if ((int64(len(data)) + offset - int64(store.headerSize)) / int64(store.batchSize)) != ((offset - int64(store.headerSize)) / int64(store.batchSize)) {
		return ErrorDataExceedBatchSize(len(data), store.batchSize, offset)
	}

	// Get current file size
	fileInfo, err := store.file.Stat()
	if err != nil {
		return ErrorFileStat(err)
	}
	fileSize := fileInfo.Size()

	n := 1 + (offset+int64(len(data))-fileSize)/int64(store.batchSize)

	// If the requested offset is beyond the current file size, allocate a new batch
	if offset+int64(len(data)) > fileSize {
		err := store.AllocateBatch(int32(n))
		if err != nil {
			return ErrorAllocatingBatch(err)
		}
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	// Write data at the given offset
	_, err = store.file.WriteAt(data, offset)
	if err != nil {
		return ErrorWritingDataAtOffset(offset, err)
	}

	return nil
}

// ReadAt reads data from the specified offset in the file
func (store *BatchStore) ReadAt(offset int64, size int32) ([]byte, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Allocate a buffer to hold the data
	data := make([]byte, size)
	// data := make([]byte, store.batchSize)

	// Read data from the given offset
	_, err := store.file.ReadAt(data, offset)
	if err != nil {
		return nil, ErrorReadingDataAtOffset(offset, err)
	}

	return data, nil
}
