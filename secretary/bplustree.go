package secretary

import (
	"fmt"
	"os"
)

// NewBPlusTree creates a new B+ Tree with the specified order
// Valid order range: [MIN_ORDER, MAX_ORDER]
func NewBPlusTree(order int) (*BPlusTree, error) {
	if order < MIN_ORDER || order > MAX_ORDER {
		return nil, fmt.Errorf("order must be between %d and %d", MIN_ORDER, MAX_ORDER)
	}

	// Initialize files with headers
	internalFile, _ := os.Create(FILE_NODE_INTERNAL)
	leafFile, _ := os.Create(FILE_NODE_LEAF)
	recordFile, _ := os.Create(FILE_RECORDS)

	return &BPlusTree{
		internalFile: internalFile,
		leafFile:     leafFile,
		recordFile:   recordFile,

		root:  nil,
		order: order,
	}, nil
}
