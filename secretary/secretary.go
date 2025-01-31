package secretary

import (
	"fmt"
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

func NewBPlusTree(collectionName string, order uint8, blockFactorPercentage uint8, nodeGroupLength uint8) (*BPlusTree, error) {
	if order < MIN_ORDER || order > MAX_ORDER {
		return nil, fmt.Errorf("order must be between %d and %d", MIN_ORDER, MAX_ORDER)
	}

	return &BPlusTree{
		collectionName: collectionName,

		blockStoreInternal: internalFile,
		blockStoreLeaf:     leafFile,
		blockStoreRecords:  recordFile,

		root:                  nil,
		order:                 order,
		blockFactorPercentage: blockFactorPercentage,
		nodeGroupLength:       nodeGroupLength,
	}, nil
}
