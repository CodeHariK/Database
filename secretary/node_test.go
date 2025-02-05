package secretary

import (
	"testing"

	"github.com/codeharik/secretary/utils/binstruct"
)

// func TestNodeSerilization(t *testing.T) {
// 	n := node{
// 		ParentOffset: 101,
// 		NextOffset:   102,
// 		PrevOffset:   103,

// 		NumKeys: 104,

// 		Keys:       [][]byte{{10, 21, 32, 34}, {110, 201, 30, 14}},
// 		KeyOffsets: []DataLocation{2, 3, 4, 5, 6},
// 	}

// 	s, err := binstruct.Serialize(n)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	var d node
// 	err = binstruct.Deserialize(s, &d)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	nJson, _ := binstruct.MarshalJSON(n)
// 	dJson, _ := binstruct.MarshalJSON(d)
// 	t.Logf("\n\n%+v\n\n%+v", n, d)
// 	t.Logf("\n\n%s\n\n%s", string(nJson), string(dJson))

// 	eq, err := binstruct.Compare(n, d)
// 	if !eq || bytes.Compare(nJson, dJson) != 0 || err != nil {
// 		t.Fatal(err)
// 	}
// }

func TestSaveRoot(t *testing.T) {
	tree := testBTree(t, "TestSaveRoot")

	tree.root = &node{
		ParentOffset: 101,
		NextOffset:   102,
		PrevOffset:   103,

		NumKeys: 104,

		// Keys:       []Key16Byte{{10, 21, 32, 34}},
		KeyOffsets: []DataLocation{2, 3, 4, 5, 6},
	}

	err := tree.saveRoot()
	if err != nil {
		t.Fatal(err)
	}

	rootBytes, err := tree.nodeBatchStore.ReadAt(SECRETARY_HEADER_LENGTH, int32(tree.nodeSize))
	if err != nil {
		t.Fatal(err)
	}

	var root node
	err = binstruct.Deserialize(rootBytes, &root)
	if err != nil {
		t.Fatal(err)
	}

	eq, err := binstruct.Compare(*tree.root, root)
	if !eq || err != nil {
		t.Fatal(err)
	}

	jsonRoot, _ := binstruct.MarshalJSON(root)
	t.Log(string(jsonRoot))
}
