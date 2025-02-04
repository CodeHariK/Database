package secretary

import (
	"bytes"
	"testing"

	"github.com/codeharik/secretary/utils/binstruct"
)

func testBTree(t *testing.T, collectionName string) *bTree {
	originalTree, err := NewBTree(
		collectionName,
		10,
		32,
		1024,
		125,
		64,
	)
	if err != nil {
		t.Fatal("NewBTree Failed")
	}
	return originalTree
}

// TestBTreeSerialization tests the serialization and deserialization of BPlusTree
func TestBTreeSerialization(t *testing.T) {
	originalTree := testBTree(t, "TestBTreeSerialization")

	// Serialize the tree
	serializedData, err := binstruct.Serialize(*originalTree)
	// serializedData, err := originalTree.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	if len(serializedData) > SECRETARY_HEADER_LENGTH {
		t.Fatalf("Expected serialized data to be %d bytes, got %d", SECRETARY_HEADER_LENGTH, len(serializedData))
	}

	var deserializedTree bTree
	err = binstruct.Deserialize(serializedData, &deserializedTree)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	eq, err := binstruct.Compare(*originalTree, deserializedTree)
	if !eq || err != nil {
		t.Fatalf("\nShould be Equal\n")
	}
}

func TestSaveHeader(t *testing.T) {
	tree := testBTree(t, "TestSaveHeader")

	// Save header
	err := tree.SaveHeader()
	if err != nil {
		t.Fatalf("SaveHeader failed: %v", err)
	}

	// Retrieve data from mock store
	savedData, err := tree.nodeBatchStore.ReadAt(0, SECRETARY_HEADER_LENGTH)
	if err != nil {
		t.Fatalf("Failed to read written data: %v", err)
	}

	// Serialize again for comparison
	expectedData, err := tree.createHeader()
	if err != nil {
		t.Fatalf("Failed to serialize expected data: %v", err)
	}

	if len(savedData) != SECRETARY_HEADER_LENGTH {
		t.Fatalf("Expected serialized data to be %d bytes, got %d", SECRETARY_HEADER_LENGTH, len(expectedData))
	}

	// Compare written data with expected serialized header
	if !bytes.Equal(savedData, expectedData) {
		t.Errorf("Saved header does not match expected serialized data")
	}
}

// func TestSearchOperations(t *testing.T) {
// 	tree, _ := NewBPlusTree(4)
// 	keys := []int{10, 5, 15, 20, 25, 30, 35}

// 	// Test search on empty tree
// 	if tree.Search(10) {
// 		t.Error("Search should return false on empty tree")
// 	}

// 	// Insert keys and verify existence
// 	for _, key := range keys {
// 		tree.Insert(key)
// 		if !tree.Search(key) {
// 			t.Errorf("Key %d should exist after insertion", key)
// 		}
// 	}

// 	// Test non-existent keys
// 	nonExistent := []int{-5, 7, 33, 100}
// 	for _, key := range nonExistent {
// 		if tree.Search(key) {
// 			t.Errorf("Key %d should not exist", key)
// 		}
// 	}

// 	// Test after deletions
// 	tree.Delete(15)
// 	if tree.Search(15) {
// 		t.Error("Deleted key 15 should not exist")
// 	}
// }

// func TestInsertAndSearch(t *testing.T) {
// 	tree, _ := NewBPlusTree(4)
// 	keys := []int{10, 5, 15, 20, 25, 30, 35}

// 	for _, key := range keys {
// 		tree.Insert(key)
// 	}

// 	// Test existing keys
// 	for _, key := range keys {
// 		if !tree.Search(key) {
// 			t.Errorf("Key %d should exist", key)
// 		}
// 	}

// 	// Test non-existent keys
// 	nonExistent := []int{-5, 7, 33, 100}
// 	for _, key := range nonExistent {
// 		if tree.Search(key) {
// 			t.Errorf("Key %d should not exist", key)
// 		}
// 	}
// }

// func TestLeafNodeOperations(t *testing.T) {
// 	tree, _ := NewBPlusTree(4)

// 	// Test single node insertion
// 	tree.Insert(10)
// 	if !tree.root.leaf || tree.root.numKeys != 1 {
// 		t.Error("Root should be single-key leaf node")
// 	}

// 	// Test leaf node split
// 	keys := []int{20, 30, 40} // Fill root leaf node
// 	for _, key := range keys {
// 		tree.Insert(key)
// 	}

// 	// After inserting 40, the leaf should split
// 	if tree.root.leaf {
// 		t.Error("Root should be internal node after split")
// 	}
// 	if tree.root.numKeys != 1 {
// 		t.Error("Root should have one key after first split")
// 	}
// }

// func TestDeleteOperations(t *testing.T) {
// 	tree, _ := NewBPlusTree(4)
// 	keys := []int{5, 10, 15, 20, 25, 30, 35, 40}
// 	for _, key := range keys {
// 		tree.Insert(key)
// 	}

// 	// Test simple deletion
// 	tree.Delete(15)
// 	if tree.Search(15) {
// 		t.Error("Deleted key 15 should not exist")
// 	}

// 	// Test deletion causing rebalance
// 	tree.Delete(5)
// 	tree.Delete(10)
// 	tree.Delete(20)

// 	remaining := []int{25, 30, 35, 40}
// 	for _, key := range remaining {
// 		if !tree.Search(key) {
// 			t.Errorf("Key %d should exist after deletions", key)
// 		}
// 	}
// }

// func TestRangeQueries(t *testing.T) {
// 	tree, _ := NewBPlusTree(4)
// 	for i := 0; i < 20; i++ {
// 		tree.Insert(i * 5) // 0, 5, 10, ..., 95
// 	}

// 	testCases := []struct {
// 		start, end int
// 		expected   []int
// 	}{
// 		{0, 100, generateSequence(0, 95, 5)},
// 		{25, 50, generateSequence(25, 50, 5)},
// 		{12, 18, []int{15}},
// 		{100, 200, []int{}},
// 		{0, 0, []int{0}},
// 	}

// 	for _, tc := range testCases {
// 		result := tree.FindRange(tc.start, tc.end)
// 		if !slices.Equal(result, tc.expected) {
// 			t.Errorf("Range [%d-%d] expected %v, got %v",
// 				tc.start, tc.end, tc.expected, result)
// 		}
// 	}
// }

// func TestLeafLinkage(t *testing.T) {
// 	tree, _ := NewBPlusTree(4)
// 	keys := []int{5, 10, 15, 20, 25, 30, 35, 40}
// 	for _, key := range keys {
// 		tree.Insert(key)
// 	}

// 	// Verify leaf order
// 	expected := []int{5, 10, 15, 20, 25, 30, 35, 40}
// 	var leaves []int
// 	current := tree.findLeaf(-1)
// 	for current != nil {
// 		leaves = append(leaves, current.keys[:current.numKeys]...)
// 		current = current.next
// 	}

// 	if !slices.Equal(leaves, expected) {
// 		t.Errorf("Leaf traversal order incorrect, got %v", leaves)
// 	}
// }

// func TestInternalNodeOperations(t *testing.T) {
// 	tree, _ := NewBPlusTree(4)

// 	// Insert enough keys to create multiple levels
// 	for i := 100; i > 0; i-- {
// 		tree.Insert(i)
// 	}

// 	// Verify root properties
// 	if tree.root.leaf || tree.root.numKeys < 1 {
// 		t.Error("Root should be internal node with multiple keys")
// 	}

// 	// Verify search in deep tree
// 	if !tree.Search(42) || tree.Search(101) {
// 		t.Error("Search in deep tree failed")
// 	}
// }

// func generateSequence(start, end, step int) []int {
// 	var seq []int
// 	for i := start; i <= end; i += step {
// 		seq = append(seq, i)
// 	}
// 	return seq
// }
