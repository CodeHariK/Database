package secretary

import (
	"bytes"
	"testing"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/binstruct"
)

func TestBTreeSerialization(t *testing.T) {
	s := dummySecretary(t)
	tree := dummyTree(t, s, 10)

	serializedData, err := binstruct.Serialize(tree)
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

	eq, err := binstruct.Compare(tree, &deserializedTree)
	if !eq || err != nil {
		t.Fatalf("\nShould be Equal\n")
	}

	s.PagerShutdown()
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
		1000,
	)
	_, invalidIncrementErr := s.NewBTree(
		"Tes",
		10,
		32,
		1024,
		225,
		1000,
	)
	_, invalidOrderErr := s.NewBTree(
		"Tes",
		210,
		32,
		1024,
		225,
		1000,
	)
	if invalidNameErr == nil || invalidIncrementErr == nil || invalidOrderErr == nil {
		t.Fatal(invalidNameErr, invalidIncrementErr, invalidOrderErr)
	}

	s.PagerShutdown()
}

func TestBtreeSaveReadHeader(t *testing.T) {
	s := dummySecretary(t)
	tree := dummyTree(t, s, 10)

	err := tree.SaveHeader()
	if err != nil {
		t.Fatalf("SaveHeader failed: %v", err)
	}

	deserializedTree, err := s.NewBTreeReadHeader(tree.CollectionName)
	if err != nil {
		t.Fatal(err)
	}

	eq, err := binstruct.Compare(tree, deserializedTree)
	if !eq || err != nil {
		t.Fatalf("\nShould be Equal\n")
	}

	tJson, tErr := binstruct.MarshalJSON(tree)
	dJson, dErr := binstruct.MarshalJSON(deserializedTree)
	if dErr != nil || tErr != nil || bytes.Compare(tJson, dJson) != 0 {
		t.Log("\n", string(tJson), "\n", string(dJson))
	}

	s.PagerShutdown()
}

func TestBTreeHeight(t *testing.T) {
	s := dummySecretary(t)
	tree := dummyTree(t, s, 4)

	var keySeq uint64 = 0

	// Set more keys to increase height
	for i := 0; i < 9; i++ {
		key := []byte(utils.GenerateSeqRandomString(&keySeq, 16, 5, 4))
		err := tree.SetKV(key, key)
		if err != nil {
			t.Fatalf("Set failed: %s", err)
		}
	}

	if tree.Height() != 2 {
		t.Fatalf("Expected height %d", tree.Height())
	}

	key := []byte(utils.GenerateSeqRandomString(&keySeq, 16, 5, 4))
	err := tree.SetKV(key, key)
	if err != nil {
		t.Fatalf("Set failed: %s", err)
	}

	if tree.Height() != 3 {
		t.Fatalf("Expected height %d", tree.Height())
	}

	s.PagerShutdown()
}

func TestBTreeClose(t *testing.T) {
	s := dummySecretary(t)

	err := s.PagerShutdown()
	if err != nil {
		t.Fatal(err)
	}
}
