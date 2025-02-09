package secretary

import (
	"reflect"
	"testing"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/binstruct"
)

func TestNewNode(t *testing.T) {
	tree, err := NewBTree(
		"TestNewNode",
		10,
		32,
		1024,
		125,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tree.NewNode(
		-1,
		-1,
		-1,
		20,
		[]DataLocation{},
		[][]byte{},
	)
	if err == nil {
		t.Fatal(err)
	}
}

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

func TestInsertNilRoot(t *testing.T) {
	tree, err := NewBTree(
		"TestInsertNilRoot",
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

	err = tree.Insert(key, value)
	if err != nil {
		t.Errorf("%s", err)
	}

	r, err := tree.Search(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}

	if r == nil {
		t.Errorf("returned nil \n")
	}

	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}
}
