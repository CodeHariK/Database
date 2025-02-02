package secretary

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// Opens or creates a file and sets up the BatchStore
func (tree *bTree) NewBatchStore(dirPath string, fileType string, level uint8) (*BatchStore, error) {
	batchSize := uint32(float64(tree.batchBaseSize) * math.Pow(float64(tree.batchIncrement)/100, float64(level)))

	path := filepath.Join(dirPath, fmt.Sprintf("%s_%d_%d.bin", fileType, level, batchSize))
	if fileType == "secretary" {
		path = filepath.Join(dirPath, fmt.Sprintf("%s.bin", fileType))
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
		file:      file,
		level:     level,
		batchSize: batchSize,
	}, nil
}

// AllocateBlock writes zeroed data in chunks of pageSize for alignment
func (store *BatchStore) AllocateBlock() (int64, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Get current file size
	fileInfo, err := store.file.Stat()
	if err != nil {
		return 0, err
	}
	offset := fileInfo.Size()

	// Align to the next page boundary
	if (uint32(offset) % store.batchSize) == 0 {
		return -1, fmt.Errorf("Error : File %d not aligned", store.level)
	}

	// Expand file by writing zeros
	zeroBuf := make([]byte, store.batchSize)
	_, err = store.file.WriteAt(zeroBuf, offset)
	if err != nil {
		return 0, err
	}

	return offset, nil
}
