package secretary

import (
	"testing"
)

// TestBTreeSerialization tests the serialization and deserialization of BPlusTree
func TestBTreeSerialization(t *testing.T) {
	originalTree, err := NewBTree(
		"Test+Collection",
		10,
		16,
		32,
		125,
		64,
		1024,
	)
	if err != nil {
		t.Fatal("NewBTree Failed")
	}

	// Serialize the tree
	serializedData, err := originalTree.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	if len(serializedData) != SECRETARY_HEADER_LENGTH {
		t.Fatalf("Expected serialized data to be %d bytes, got %d", SECRETARY_HEADER_LENGTH, len(serializedData))
	}

	// Deserialize the tree
	deserializedTree, err := DeserializeBPlusTree(serializedData)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Compare fields to ensure correctness
	if deserializedTree.order != originalTree.order {
		t.Errorf("Expected order %d, got %d", originalTree.order, deserializedTree.order)
	}
	if deserializedTree.keySize != originalTree.keySize {
		t.Errorf("Expected keySize %d, got %d", originalTree.keySize, deserializedTree.keySize)
	}
	if deserializedTree.batchNumLevel != originalTree.batchNumLevel {
		t.Errorf("Expected batchNumLevel %d, got %d", originalTree.batchNumLevel, deserializedTree.batchNumLevel)
	}
	if deserializedTree.batchBaseSize != originalTree.batchBaseSize {
		t.Errorf("Expected batchBaseSize %d, got %d", originalTree.batchBaseSize, deserializedTree.batchBaseSize)
	}
	if deserializedTree.batchIncrement != originalTree.batchIncrement {
		t.Errorf("Expected batchIncrement %d, got %d", originalTree.batchIncrement, deserializedTree.batchIncrement)
	}
	if deserializedTree.batchLength != originalTree.batchLength {
		t.Errorf("Expected batchLength %d, got %d", originalTree.batchLength, deserializedTree.batchLength)
	}

	// Check collectionName (trim any extra null bytes)
	expectedCollectionName := originalTree.collectionName
	actualCollectionName := deserializedTree.collectionName

	if actualCollectionName != expectedCollectionName {
		t.Errorf("Expected collectionName '%s', got '%s'", expectedCollectionName, actualCollectionName)
	}
}
