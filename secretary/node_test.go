package secretary

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/binstruct"
)

func TestSaveRoot(t *testing.T) {
	tree, err := NewBTree(
		"TestSaveRoot",
		10,
		32,
		1024,
		125,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	tree.root = &Node{
		ParentOffset: 101,
		NextOffset:   102,
		PrevOffset:   103,

		NumKeys: 104,

		Keys:       [][]byte{{10, 21, 32, 34, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		KeyOffsets: []DataLocation{2, 3, 4, 5, 6},
	}

	err = tree.saveRoot()
	if err != nil {
		t.Fatal(err)
	}

	rootBytes, err := tree.nodeBatchStore.ReadAt(SECRETARY_HEADER_LENGTH, int32(tree.nodeSize))
	if err != nil {
		t.Fatal(err)
	}

	var root Node
	err = binstruct.Deserialize(rootBytes, &root)
	if err != nil {
		t.Fatal(err)
	}

	jsonS, _ := binstruct.MarshalJSON(tree.root)
	jsonD, _ := binstruct.MarshalJSON(root)

	t.Log("\n", *tree.root, "\n", root, "\n", string(jsonS), "\n", string(jsonD), "\n")

	eq, err := binstruct.Compare(*tree.root, root)
	if !eq || err != nil {
		t.Fatal(err)
	}
}

func TestInsert(t *testing.T) {
	tree, err := NewBTree(
		"TestInsert",
		10,
		32,
		1024,
		125,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	r, err := tree.Search([]byte(utils.GenerateRandomString(16)))
	if err == nil || r != nil {
		t.Error("expected error and got nil", err, r)
	}

	key := []byte(utils.GenerateRandomString(16))
	value := []byte("Hello world!")
	err = tree.Insert(key, value)
	if err != nil {
		t.Errorf("%s", err)
	}

	r, err = tree.Search(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}

	if r == nil {
		t.Errorf("returned nil \n")
	}

	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}

	err = tree.Insert(key, append(value, []byte("world1")...))
	if err == nil {
		t.Errorf("expected error but got nil %v", err)
	}

	r, err = tree.Search(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}
	if r == nil {
		t.Errorf("returned nil \n")
	}

	if bytes.Compare(r.Value, value) != 0 {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}

	if tree.root.NumKeys != 1 {
		t.Errorf("expected 1 key and got %d", tree.root.NumKeys)
	}
}

func TestDelete(t *testing.T) {
	tree, err := NewBTree(
		"TestInsert",
		8,
		32,
		1024,
		125,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	key := []byte(utils.GenerateRandomString(16))
	value := []byte("Hello world!")

	err = tree.Delete(key)
	if err == nil {
		t.Errorf("expected error and got nil")
	}

	err = tree.Insert(key, value)
	if err != nil {
		t.Error(err)
	}

	r, err := tree.Search(key)
	if err != nil {
		t.Error(err)
	}
	if r == nil {
		t.Errorf("returned nil \n")
	}
	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}

	err = tree.Delete(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}

	r, err = tree.Search(key)
	if err == nil {
		t.Error("expected error and got nil", err)
	}
	if r != nil {
		t.Errorf("returned struct after delete \n")
	}

	multipleKeys := make([][]byte, 40)
	for i := range multipleKeys {
		multipleKeys[i] = []byte(utils.GenerateSeqRandomString(16, 4))
		err = tree.Insert(multipleKeys[i], multipleKeys[i])
		if err != nil {
			t.Errorf("Insert failed: %s", err)
		}
	}
	// for i := range multipleKeys {
	// 	r, err = tree.Search(multipleKeys[i])
	// 	if err != nil || bytes.Compare(r.Value, multipleKeys[i]) != 0 {
	// 		t.Errorf("Search failed: %d : %s", i, err)
	// 	}
	// }
	// for i := range multipleKeys {
	// 	err = tree.Delete(multipleKeys[i])
	// 	if err != nil {
	// 		t.Errorf("Delete failed: %s", err)
	// 	}
	// }
}

func TestUpdate(t *testing.T) {
	tree, err := NewBTree(
		"TestInsert",
		10,
		32,
		1024,
		125,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	key := []byte(utils.GenerateRandomString(16))
	value := []byte("Hello world!")

	err = tree.Update(key, value)
	if err == nil {
		t.Error(err)
	}

	err = tree.Insert(key, value)
	if err != nil {
		t.Error(err)
	}
	r, err := tree.Search(key)
	if err != nil {
		t.Error(err)
	}
	if r == nil {
		t.Errorf("returned nil \n")
	}
	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}

	newValue := []byte("Alola world!")
	err = tree.Update(key, newValue)
	if err != nil {
		t.Error(err)
	}
	r, err = tree.Search(key)
	if err != nil {
		t.Error(err)
	}
	if r == nil {
		t.Errorf("returned nil \n")
	}
	if !reflect.DeepEqual(r.Value, newValue) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}
}
