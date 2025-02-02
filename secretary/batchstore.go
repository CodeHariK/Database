package secretary

import (
	"fmt"
	"math"
	"os"
)

// Opens or creates a file and sets up the BatchStore
func (tree *bTree) NewBatchStore(fileType string, level uint8) (*BatchStore, error) {
	batchSize := uint32(float64(tree.batchBaseSize) * math.Pow(float64(tree.batchIncrement)/100, float64(level)))

	headerSize := 0

	path := fmt.Sprintf("SECRETARY/%s/%s_%d_%d.bin", tree.collectionName, fileType, level, batchSize)
	if fileType == "index" {

		headerSize = SECRETARY_HEADER_LENGTH

		path = fmt.Sprintf("SECRETARY/%s/%s.bin", tree.collectionName, fileType)

	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		file.Close()
		fmt.Printf("Create File : %s\n", path)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Open File : %s\n", path)

	return &BatchStore{
		file:       file,
		level:      level,
		headerSize: uint8(headerSize),
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
		return fmt.Errorf("Error : File %s not aligned", fileInfo.Name())
	}

	// Expand file by writing zeros
	zeroBuf := make([]byte, store.batchSize*uint32(numBatch))
	_, err = store.file.WriteAt(zeroBuf, fileSize)
	if err != nil {
		return err
	}

	return nil
}

// WriteAtOffset writes data at the specified offset in the file.
// If there is not enough free space, it allocates a new batch.
func (store *BatchStore) WriteAtOffset(offset int64, data []byte) error {
	// Ensure data size does not exceed batchSize
	if ((int64(len(data)) + offset) / int64(store.batchSize)) != (offset / int64(store.batchSize)) {
		return fmt.Errorf("Error: Data size %d exceeds batch size %d at offset %d", len(data), store.batchSize, offset)
	}

	// Get current file size
	fileInfo, err := store.file.Stat()
	if err != nil {
		return fmt.Errorf("Error getting file size: %v", err)
	}
	fileSize := fileInfo.Size()

	n := 1 + (offset+int64(len(data))-fileSize)/int64(store.batchSize)

	// If the requested offset is beyond the current file size, allocate a new batch
	if offset+int64(len(data)) > fileSize {
		err := store.AllocateBatch(int32(n))
		if err != nil {
			return fmt.Errorf("Error allocating batch: %v", err)
		}
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	// Write data at the given offset
	_, err = store.file.WriteAt(data, offset)
	if err != nil {
		return fmt.Errorf("Error writing data at offset %d: %v", offset, err)
	}

	return nil
}

// ReadAtOffset reads data from the specified offset in the file
func (store *BatchStore) ReadAtOffset(offset int64, size int32) ([]byte, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Allocate a buffer to hold the data
	data := make([]byte, size)
	// data := make([]byte, store.batchSize)

	// Read data from the given offset
	_, err := store.file.ReadAt(data, offset)
	if err != nil {
		return nil, fmt.Errorf("Error reading data at offset %d: %v", offset, err)
	}

	return data, nil
}
