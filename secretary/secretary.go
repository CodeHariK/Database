package secretary

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/codeharik/secretary/utils"
)

func new() *Secretary {
	secretary := &Secretary{
		tree: map[string]*BPlusTree{
			"users":    {collectionName: "users"},
			"products": {collectionName: "products"},
		},
	}

	basePath := "data"
	err := secretary.InitializeStorage(basePath)
	if err != nil {
		fmt.Println("Error initializing storage:", err)
	} else {
		fmt.Println("Storage initialized successfully")
	}

	return secretary
}

// Create necessary files inside a directory
func createFiles(basePath string) error {
	files := []string{"internal.bin", "leaf.bin"}

	// Create internal and leaf files
	for _, fileName := range files {
		filePath := filepath.Join(basePath, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			file, err := os.Create(filePath)
			if err != nil {
				return err
			}
			file.Close()
		}
	}

	// Create 32 index_records.bin files
	for i := 1; i <= 32; i++ {
		filePath := filepath.Join(basePath, fmt.Sprintf("records_%d.bin", i))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			file, err := os.Create(filePath)
			if err != nil {
				return err
			}
			file.Close()
		}
	}

	return nil
}

// Initialize directories and files for all collections
func (s *Secretary) InitializeStorage(basePath string) error {
	for collectionName := range s.tree {
		collectionPath := filepath.Join(basePath, collectionName)

		// Ensure directory exists
		if err := utils.EnsureDir(collectionPath); err != nil {
			return err
		}

		// Create necessary files
		if err := createFiles(collectionPath); err != nil {
			return err
		}
	}

	return nil
}

/*
InternalBlockStore

------512 bytes------
SECRETARY				(9 bytes)
len(collectionName)		(uint8)
collectionName			(string)
order 					(uint8)
blockFactorPercentage	(uint8)
nodeGroupLength 		(uint8)
---------------------
1  			Root Internal
order		Internal
order ^ 2	Internal
order ^ 3	Internal
order ^ 4	Internal
...
order ^ n	Leaf
---------------------
*/

const (
	MIN_ORDER = uint8(3)  // Minimum allowed order for the B+ Tree
	MAX_ORDER = uint8(40) // Maximum allowed order for the B+ Tree
)

func NewBPlusTree(
	collectionName string, order uint8, numLevel uint8,
	batchNumLevel uint8, batchIncrement uint8, batchLength uint8, batchBaseSize uint32,
) (*BPlusTree, error) {
	if order < MIN_ORDER || order > MAX_ORDER {
		log.Fatalf("order must be between %d and %d", MIN_ORDER, MAX_ORDER)
	}

	if batchIncrement < 110 || batchIncrement > 200 {
		log.Fatalf("batchSizePercentage must be between 110 and 200")
	}

	safeCollectionName := utils.RemoveNonAlphanumeric(collectionName)
	if safeCollectionName == "" {
		log.Fatalf("collection name not valid, should be a-z 0-9")
	}

	if dirNotExist := utils.EnsureDir(safeCollectionName); dirNotExist != nil {
		log.Fatalf(safeCollectionName, dirNotExist.Error())
	}

	tree := &BPlusTree{
		collectionName: safeCollectionName,

		root:  nil,
		order: order,

		batchNumLevel:  batchNumLevel,
		batchBaseSize:  batchBaseSize,
		batchIncrement: batchIncrement,
		batchLength:    batchLength,
	}

	batchStores := make([]*BatchStore, numLevel)
	for i := range batchStores {
		store, err := tree.NewBatchStore(safeCollectionName, 0)
		if err != nil {
			log.Fatal(err.Error())
		}

		batchStores[i] = store
	}

	return tree, nil
}

// Opens or creates a file and sets up the BatchStore
func (tree *BPlusTree) NewBatchStore(filePath string, level uint8) (*BatchStore, error) {
	size := uint32(tree.batchBaseSize)*uint32(tree.batchIncrement/100) ^ uint32(level)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	return &BatchStore{
		file:  file,
		level: level,
		size:  size,
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
	if (uint32(offset) % store.size) == 0 {
		log.Fatalf("Error : File %d not aligned", store.level)
	}

	// Expand file by writing zeros
	zeroBuf := make([]byte, store.size)
	_, err = store.file.WriteAt(zeroBuf, offset)
	if err != nil {
		return 0, err
	}

	return offset, nil
}
