package secretary

import (
	"bytes"
	"testing"

	"github.com/codeharik/secretary/utils"
)

func testBTree(t *testing.T, collectionName string) *bTree {
	originalTree, err := NewBTree(
		collectionName,
		10,
		16,
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
	serializedData, err := utils.SerializeBinaryStruct(*originalTree)
	// serializedData, err := originalTree.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	if len(serializedData) > SECRETARY_HEADER_LENGTH {
		t.Fatalf("Expected serialized data to be %d bytes, got %d", SECRETARY_HEADER_LENGTH, len(serializedData))
	}

	var deserializedTree bTree
	err = utils.DeserializeBinaryStruct(serializedData, &deserializedTree)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	eq, err := utils.CompareBinaryStruct(*originalTree, deserializedTree)
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
