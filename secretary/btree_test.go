package secretary

import (
	"bytes"
	"testing"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/binstruct"
)

// TestBTreeSerialization tests the serialization and deserialization of BPlusTree
func TestBTreeSerialization(t *testing.T) {
	_, originalTree := dummyTree(t, "TestBTreeSerialization", 10)

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
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	_, invalidNameErr := s.NewBTree(
		"Tes",
		10,
		32,
		1024,
		125,
		10,
	)
	_, invalidIncrementErr := s.NewBTree(
		"Tes",
		10,
		32,
		1024,
		225,
		10,
	)
	_, invalidOrderErr := s.NewBTree(
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
	s, tree := dummyTree(t, "TestSaveHeader", 10)

	err := tree.SaveHeader()
	if err != nil {
		t.Fatalf("SaveHeader failed: %v", err)
	}

	deserializedTree, err := s.NewBTreeReadHeader(tree.CollectionName)
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

func TestBTreeHeight(t *testing.T) {
	_, tree := dummyTree(t, "TestSaveHeader", 4)

	var keySeq uint64 = 0

	// Set more keys to increase height
	for i := 0; i < 9; i++ {
		key := []byte(utils.GenerateSeqRandomString(&keySeq, 16, 4))
		err := tree.Set(key, key)
		if err != nil {
			t.Errorf("Set failed: %s", err)
		}
	}

	if tree.Height() != 2 {
		t.Errorf("Expected height %d", tree.Height())
	}

	key := []byte(utils.GenerateSeqRandomString(&keySeq, 16, 4))
	err := tree.Set(key, key)
	if err != nil {
		t.Errorf("Set failed: %s", err)
	}

	if tree.Height() != 3 {
		t.Errorf("Expected height %d", tree.Height())
	}

	jsonOutput, err := tree.MarshalGraphJSON()
	if err != nil {
		t.Fatal("Error:", err)
		return
	}
	t.Log(tree.Order, string(jsonOutput))
}

func TestBTreeClose(t *testing.T) {
	_, tree := dummyTree(t, "TestSaveHeader", 4)

	err := tree.close()
	if err != nil {
		t.Fatal(err)
	}
}
