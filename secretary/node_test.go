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
		t.Fatalf("%s", err)
	}

	r, err = tree.Search(key)
	if err != nil {
		t.Fatalf("%s\n", err)
	}
	if r == nil || !reflect.DeepEqual(r.Value, value) {
		t.Fatalf("expected %v and got %v \n", value, r)
	}

	// Duplicate Key error
	err = tree.Insert(key, append(value, []byte("world1")...))
	if err == nil {
		t.Fatalf("expected error but got nil %v", err)
	}

	r, err = tree.Search(key)
	if err != nil {
		t.Fatalf("%s\n", err)
	}
	if r == nil || bytes.Compare(r.Value, value) != 0 {
		t.Fatalf("expected %v and got %v \n", value, r)
	}
}

func TestNodeScan(t *testing.T) {
	tests := []struct {
		keys     [][]byte
		search   []byte
		expected int
	}{
		// Search for existing keys
		{
			[][]byte{[]byte("a"), []byte("c"), []byte("e")},
			[]byte("c"),
			1,
		},
		{
			[][]byte{[]byte("apple"), []byte("banana"), []byte("cherry")},
			[]byte("banana"),
			1,
		},

		// Search for non-existing keys (returns insertion index)
		{
			[][]byte{[]byte("b"), []byte("c"), []byte("e")},
			[]byte("a"),
			0,
		},
		{
			[][]byte{[]byte("a"), []byte("c"), []byte("e")},
			[]byte("b"),
			1,
		},
		{
			[][]byte{[]byte("a"), []byte("c"), []byte("e")},
			[]byte("f"),
			3,
		},
		{
			[][]byte{[]byte("apple"), []byte("banana"), []byte("cherry")},
			[]byte("blueberry"),
			2,
		},

		// Edge cases
		{
			[][]byte{},
			[]byte("z"),
			0,
		}, // Empty node
		{
			[][]byte{[]byte("m")},
			[]byte("m"),
			0,
		}, // Single key (exact match)
		{
			[][]byte{[]byte("m")},
			[]byte("a"),
			0,
		}, // Single key (less than)
		{
			[][]byte{[]byte("m")},
			[]byte("z"),
			1,
		}, // Single key (greater than)
	}

	for _, test := range tests {
		node := &Node{Keys: test.keys}
		result := node.nodeScan(test.search)
		if result != test.expected {
			t.Errorf("nodeSearch(%q) = %d, expected %d", test.search, result, test.expected)
		}
	}
}

func TestSearchLeafNode(t *testing.T) {
	// Create a simple B+ tree manually
	root := &Node{
		Keys: [][]byte{[]byte("m")},
		children: []*Node{
			{
				Keys: [][]byte{[]byte("a"), []byte("e"), []byte("h")},
			}, // Left child
			{
				Keys: [][]byte{[]byte("n"), []byte("r"), []byte("z")},
			}, // Right child
		},
	}

	tree := &BTree{root: root}

	tests := []struct {
		key      []byte
		expected *Node
	}{
		{
			[]byte("b"),
			root.children[0],
		}, // Should go to left child
		{
			[]byte("g"),
			root.children[0],
		}, // Should go to left child
		{
			[]byte("q"),
			root.children[1],
		}, // Should go to right child
		{
			[]byte("z"),
			root.children[1],
		}, // Should go to right child
	}

	for _, test := range tests {
		result := tree.searchLeafNode(test.key)
		if result != test.expected {
			t.Errorf("searchLeafNode(%q) returned wrong leaf node", test.key)
		}
	}

	// Test empty tree
	emptyTree := &BTree{root: &Node{}}
	if emptyTree.searchLeafNode([]byte("x")) != emptyTree.root {
		t.Errorf("searchLeafNode on empty tree should return root")
	}

	// Test single-node tree
	singleNodeTree := &BTree{root: &Node{Keys: [][]byte{[]byte("a"), []byte("b"), []byte("c")}}}
	if singleNodeTree.searchLeafNode([]byte("b")) != singleNodeTree.root {
		t.Errorf("searchLeafNode on single-node tree should return root")
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
		t.Fatalf("expected error and got nil")
	}

	err = tree.Insert(key, value)
	if err != nil {
		t.Error(err)
	}

	r, err := tree.Search(key)
	if err != nil {
		t.Error(err)
	}
	if r == nil || !reflect.DeepEqual(r.Value, value) {
		t.Fatalf("expected %v and got %v \n", value, r)
	}

	err = tree.Delete(key)
	if err != nil {
		t.Fatalf("%s\n", err)
	}

	r, err = tree.Search(key)
	if r != nil || err == nil {
		t.Error("expected error and got struct", err)
	}

	multipleKeys := make([][]byte, 40)
	for i := range multipleKeys {
		multipleKeys[i] = []byte(utils.GenerateSeqRandomString(16, 4))
		err = tree.Insert(multipleKeys[i], multipleKeys[i])
		if err != nil {
			t.Fatalf("Insert failed: %s", err)
		}
	}
	// for i := range multipleKeys {
	// 	r, err = tree.Search(multipleKeys[i])
	// 	if err != nil || bytes.Compare(r.Value, multipleKeys[i]) != 0 {
	// 		t.Fatalf("Search failed: %d : %s", i, err)
	// 	}
	// }
	// for i := range multipleKeys {
	// 	err = tree.Delete(multipleKeys[i])
	// 	if err != nil {
	// 		t.Fatalf("Delete failed: %s", err)
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
	if r == nil || !reflect.DeepEqual(r.Value, value) {
		t.Fatalf("expected %v and got %v \n", value, r)
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
	if r == nil || !reflect.DeepEqual(r.Value, newValue) {
		t.Fatalf("expected %v and got %v \n", value, r)
	}
}
