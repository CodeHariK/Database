package secretary

import (
	"bytes"
	"testing"

	"github.com/codeharik/secretary/utils/binstruct"
)

// TestBTreeSerialization tests the serialization and deserialization of BPlusTree
func TestBTreeSerialization(t *testing.T) {
	originalTree, err := NewBTree(
		"TestBTreeSerialization",
		10,
		32,
		1024,
		125,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Serialize the tree
	serializedData, err := binstruct.Serialize(*originalTree)
	// serializedData, err := originalTree.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	if len(serializedData) > SECRETARY_HEADER_LENGTH {
		t.Fatalf("Expected serialized data to be %d bytes, got %d", SECRETARY_HEADER_LENGTH, len(serializedData))
	}

	var deserializedTree BTree
	err = binstruct.Deserialize(serializedData, &deserializedTree)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	eq, err := binstruct.Compare(*originalTree, deserializedTree)
	if !eq || err != nil {
		t.Fatalf("\nShould be Equal\n")
	}
}

func TestBtreeInvalid(t *testing.T) {
	_, invalidNameErr := NewBTree(
		"Tes",
		10,
		32,
		1024,
		125,
		10,
	)
	_, invalidIncrementErr := NewBTree(
		"Tes",
		10,
		32,
		1024,
		225,
		10,
	)
	_, invalidOrderErr := NewBTree(
		"Tes",
		210,
		32,
		1024,
		225,
		10,
	)
	if invalidNameErr == nil || invalidIncrementErr == nil || invalidOrderErr == nil {
		t.Fatal(invalidNameErr, invalidIncrementErr, invalidOrderErr)
	}
}

func TestSaveReadHeader(t *testing.T) {
	tree, err := NewBTree(
		"TestSaveHeader",
		10,
		32,
		1024,
		125,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = tree.SaveHeader()
	if err != nil {
		t.Fatalf("SaveHeader failed: %v", err)
	}

	deserializedTree, err := NewBTreeReadHeader(tree.CollectionName)
	if err != nil {
		t.Fatal(err)
	}

	eq, err := binstruct.Compare(*tree, *deserializedTree)
	if !eq || err != nil {
		t.Fatalf("\nShould be Equal\n")
	}

	tJson, tErr := binstruct.MarshalJSON(*tree)
	dJson, dErr := binstruct.MarshalJSON(deserializedTree)
	if dErr != nil || tErr != nil || bytes.Compare(tJson, dJson) != 0 {
		t.Log("\n", string(tJson), "\n", string(dJson))
	}
}

// func TestBTreeHeight(t *testing.T) {
// 	tree := &BTree{}

// 	// Insert first key-value pair (root node only)
// 	key1 := []byte(utils.GenerateRandomString(16))
// 	value1 := []byte("Hello world!")
// 	err := tree.Insert(key1, value1)
// 	if err != nil {
// 		t.Errorf("Insert failed: %s", err)
// 	}
// 	if tree.Height() != 1 {
// 		t.Errorf("Expected height 1, got %d", tree.Height())
// 	}

// 	// Insert more keys to increase height
// 	for i := 0; i < 3; i++ {
// 		key := []byte(utils.GenerateRandomString(16))
// 		value := []byte(fmt.Sprintf("value : %d", i))
// 		err = tree.Insert(key, value)
// 		if err != nil {
// 			t.Errorf("Insert failed: %s", err)
// 		}
// 	}

// 	// Check if the height increased after multiple insertions
// 	if tree.Height() != 3 {
// 		t.Errorf("Expected height %d", tree.Height())
// 	}

// 	jsonOutput, err := tree.ConvertBTreeToJSON()
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}
// 	t.Log(tree.Order, jsonOutput)
// }
