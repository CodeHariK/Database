package secretary

// import (
// 	"slices"
// 	"testing"
// )

// func TestNewBPlusTree(t *testing.T) {
// 	// Test valid order
// 	_, err := NewBPlusTree(4)
// 	if err != nil {
// 		t.Errorf("Expected no error, got %v", err)
// 	}

// 	// Test invalid orders
// 	_, err = NewBPlusTree(2)
// 	if err == nil {
// 		t.Error("Expected error for order < MIN_ORDER")
// 	}

// 	_, err = NewBPlusTree(21)
// 	if err == nil {
// 		t.Error("Expected error for order > MAX_ORDER")
// 	}
// }

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
