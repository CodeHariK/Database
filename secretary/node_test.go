package secretary

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/binstruct"
)

func dummyTree(t *testing.T, collectionName string, order uint8) (*Secretary, *BTree) {
	s, serr := New()
	tree, err := s.NewBTree(
		collectionName,
		order,
		32,
		1024,
		125,
		10,
		1000,
	)
	if serr != nil || err != nil {
		t.Fatal(err)
	}
	return s, tree
}

func TestSaveRoot(t *testing.T) {
	_, tree := dummyTree(t, "TestSaveRoot", 10)

	root := Node{
		ParentOffset: 101,
		NextOffset:   102,
		PrevOffset:   103,

		Keys:       [][]byte{{10, 21, 32, 34, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		KeyOffsets: []DataLocation{2, 3, 4, 5, 6},
	}
	tree.root = &root

	err := tree.saveRoot()
	if err != nil {
		t.Fatal(err)
	}

	err = tree.readRoot()
	if err != nil {
		t.Fatal(err)
	}

	jsonS, _ := binstruct.MarshalJSON(tree.root)
	jsonD, _ := binstruct.MarshalJSON(&root)

	t.Log("\n", tree.root, "\n", &root, "\n", string(jsonS), "\n", string(jsonD), "\n")

	eq, err := binstruct.Compare(tree.root, &root)
	if !eq || err != nil {
		t.Fatal(err)
	}
}

func TestNodeScan(t *testing.T) {
	tests := []struct {
		keys          [][]byte
		search        []byte
		expectedIndex int
		expectedFound bool
	}{
		// Search for existing keys
		{
			[][]byte{[]byte("a"), []byte("c"), []byte("e")},
			[]byte("c"),
			1,
			true,
		},
		{
			[][]byte{[]byte("apple"), []byte("banana"), []byte("cherry")},
			[]byte("banana"),
			1,
			true,
		},

		// Search for non-existing keys (returns set index)
		{
			[][]byte{[]byte("b"), []byte("c"), []byte("e")},
			[]byte("a"),
			0,
			false,
		},
		{
			[][]byte{[]byte("a"), []byte("c"), []byte("e")},
			[]byte("b"),
			1,
			false,
		},
		{
			[][]byte{[]byte("a"), []byte("c"), []byte("e")},
			[]byte("f"),
			3,
			false,
		},
		{
			[][]byte{[]byte("apple"), []byte("banana"), []byte("cherry")},
			[]byte("blueberry"),
			2,
			false,
		},

		// Edge cases
		{
			[][]byte{},
			[]byte("z"),
			0,
			false,
		}, // Empty node
		{
			[][]byte{[]byte("m")},
			[]byte("m"),
			0,
			true,
		}, // Single key (exact match)
		{
			[][]byte{[]byte("m")},
			[]byte("a"),
			0,
			false,
		}, // Single key (less than)
		{
			[][]byte{[]byte("m")},
			[]byte("z"),
			1,
			false,
		}, // Single key (greater than)
	}

	for _, test := range tests {
		node := &Node{Keys: test.keys}
		result, found := node.getKey(test.search)
		if result != test.expectedIndex || found != test.expectedFound {
			t.Fatalf("NodeScan(%q) = %d, expected %d", test.search, result, test.expectedIndex)
		}
	}
}

func TestGetLeafNode(t *testing.T) {
	// Create a simple B+ tree manually
	root := &Node{
		Keys: [][]byte{[]byte("h"), []byte("r")},
		children: []*Node{
			{
				Keys: [][]byte{[]byte("b"), []byte("e")},
			},
			{
				Keys: [][]byte{[]byte("h"), []byte("k"), []byte("o")},
			},
			{
				Keys: [][]byte{[]byte("r"), []byte("u"), []byte("y")},
			},
		},
	}

	// abcdefg   hijklmnopq	  rstuvwxyz
	// a_cd_fg   +ij_lmn_pq	  +st_vwx_z
	// 0011122   0111222233	  011122223

	tree := &BTree{root: root}

	tests := []struct {
		key           []byte
		expectedNode  *Node
		expectedIndex int
		expectedFound bool
	}{
		{[]byte("a"), root.children[0], 0, false}, // Before "b"
		{[]byte("b"), root.children[0], 0, true},  // Exists
		{[]byte("c"), root.children[0], 1, false}, // Between "b" and "e"
		{[]byte("d"), root.children[0], 1, false}, // Between "b" and "e"
		{[]byte("e"), root.children[0], 1, true},  // Exists
		{[]byte("f"), root.children[0], 2, false}, // Between "e" and "h"
		{[]byte("g"), root.children[0], 2, false}, // Between "e" and "h"

		{[]byte("h"), root.children[1], 0, true},  // Exists
		{[]byte("i"), root.children[1], 1, false}, // Between "h" and "k"
		{[]byte("j"), root.children[1], 1, false}, // Between "h" and "k"
		{[]byte("k"), root.children[1], 1, true},  // Exists
		{[]byte("l"), root.children[1], 2, false}, // Between "k" and "o"
		{[]byte("m"), root.children[1], 2, false}, // Between "k" and "o"
		{[]byte("n"), root.children[1], 2, false}, // Between "k" and "o"
		{[]byte("o"), root.children[1], 2, true},  // Exists
		{[]byte("p"), root.children[1], 3, false}, // Between "o" and "r"
		{[]byte("q"), root.children[1], 3, false}, // Between "o" and "r"

		{[]byte("r"), root.children[2], 0, true},  // Exists
		{[]byte("s"), root.children[2], 1, false}, // Between "r" and "u"
		{[]byte("t"), root.children[2], 1, false}, // Between "r" and "u"
		{[]byte("u"), root.children[2], 1, true},  // Exists
		{[]byte("v"), root.children[2], 2, false}, // Between "u" and "y"
		{[]byte("w"), root.children[2], 2, false}, // Between "u" and "y"
		{[]byte("x"), root.children[2], 2, false}, // Between "u" and "y"
		{[]byte("y"), root.children[2], 2, true},  // Exists
		{[]byte("z"), root.children[2], 3, false}, // After "y"
	}

	for _, test := range tests {
		result, index, found := tree.getLeafNode(test.key)

		if result.NodeID != test.expectedNode.NodeID || index != test.expectedIndex || found != test.expectedFound {
			t.Fatalf("GetLeafNode(%q) returned wrong leaf node\n ExpNode:%d - Got:%d\n ExpID:%d - Got:%d\n ExpFound:%v - Got:%v\n %v",
				test.key,
				test.expectedNode.NodeID, result.NodeID,
				test.expectedIndex, index,
				test.expectedFound, found,
				utils.ArrayToStrings(result.Keys),
			)
		}
	}

	// Test empty tree
	emptyTree := &BTree{root: &Node{}}
	emptyNode, _, _ := emptyTree.getLeafNode([]byte("x"))
	if emptyNode.NodeID != emptyTree.root.NodeID {
		t.Fatalf("GetLeafNode on empty tree should return root")
	}

	// Test single-node tree
	singleNodeTree := &BTree{root: &Node{Keys: [][]byte{[]byte("a"), []byte("b"), []byte("c")}}}
	singleNode, _, _ := singleNodeTree.getLeafNode([]byte("b"))
	if singleNode.NodeID != singleNodeTree.root.NodeID {
		t.Fatalf("GetLeafNode on single-node tree should return root")
	}
}

func TestSet(t *testing.T) {
	_, tree := dummyTree(t, "TestSet", 10)

	r, err := tree.Get([]byte(utils.GenerateRandomString(16)))
	if err == nil || r != nil {
		t.Error("expected error and got nil", err, r)
	}

	key := []byte(utils.GenerateRandomString(16))
	value := []byte("Hello world!")
	err = tree.Set(key, value)
	if err != nil {
		t.Fatalf("%s", err)
	}

	r, err = tree.Get(key)
	if err != nil {
		t.Fatalf("%s\n", err)
	}
	if r == nil || !reflect.DeepEqual(r.Value, value) {
		t.Fatalf("expected %v and got %v \n", value, r)
	}

	// Duplicate Key error
	err = tree.Set(key, append(value, []byte("world1")...))
	if err == nil {
		t.Fatalf("expected error but got nil %v", err)
	}

	r, err = tree.Get(key)
	if err != nil {
		t.Fatalf("%s\n", err)
	}
	if r == nil || bytes.Compare(r.Value, value) != 0 {
		t.Fatalf("expected %v and got %v \n", value, r)
	}
}

func TestUpdate(t *testing.T) {
	_, tree := dummyTree(t, "TestUpdate", 10)

	key := []byte(utils.GenerateRandomString(16))
	value := []byte("Hello world!")

	err := tree.Update(key, value)
	if err == nil {
		t.Error(err)
	}

	err = tree.Set(key, value)
	if err != nil {
		t.Error(err)
	}
	r, err := tree.Get(key)
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
	r, err = tree.Get(key)
	if err != nil {
		t.Error(err)
	}
	if r == nil || !reflect.DeepEqual(r.Value, newValue) {
		t.Fatalf("expected %v and got %v \n", value, r)
	}
}

func TestRangeScan(t *testing.T) {
	node1 := &Node{
		NodeID:  2,
		Keys:    [][]byte{[]byte("b"), []byte("e")},
		records: []*Record{{Value: []byte("b")}, {Value: []byte("e")}},
	}
	node2 := &Node{
		NodeID:  3,
		Keys:    [][]byte{[]byte("h"), []byte("k"), []byte("o")},
		records: []*Record{{Value: []byte("h")}, {Value: []byte("k")}, {Value: []byte("o")}},
	}
	node3 := &Node{
		NodeID:  4,
		Keys:    [][]byte{[]byte("r"), []byte("u"), []byte("y")},
		records: []*Record{{Value: []byte("r")}, {Value: []byte("u")}, {Value: []byte("y")}},
	}

	node1.next = node2
	node2.prev = node1
	node2.next = node3
	node3.prev = node2

	root := &Node{
		NodeID: 1,
		Keys:   [][]byte{[]byte("h"), []byte("r")},
		children: []*Node{
			node1,
			node2,
			node3,
		},
	}

	// abcdefg   hijklmnopq	  rstuvwxyz
	// a_cd_fg   +ij_lmn_pq	  +st_vwx_z
	// 0011122   0111222233	  011122223

	tree := &BTree{root: root}

	tests := []struct {
		startKey []byte
		endKey   []byte
		expected []string
	}{
		{[]byte("a"), []byte("z"), []string{"b", "e", "h", "k", "o", "r", "u", "y"}},
		{[]byte("b"), []byte("z"), []string{"b", "e", "h", "k", "o", "r", "u", "y"}},
		{[]byte("c"), []byte("x"), []string{"e", "h", "k", "o", "r", "u"}},
		{[]byte("c"), []byte("y"), []string{"e", "h", "k", "o", "r", "u", "y"}},
		{[]byte("c"), []byte("z"), []string{"e", "h", "k", "o", "r", "u", "y"}},
		{[]byte("c"), []byte("c"), []string{}},
		{[]byte("e"), []byte("e"), []string{"e"}},
		{[]byte("l"), []byte("m"), []string{}},
	}

	for _, tt := range tests {
		results := tree.RangeScan(tt.startKey, tt.endKey)
		var resultKeys []string
		for _, record := range results {
			resultKeys = append(resultKeys, string(record.Value))
		}
		if !utils.CompareArray(resultKeys, tt.expected) {
			t.Fatalf("RangeScan(%q, %q) = %v, expected %v", tt.startKey, tt.endKey, resultKeys, tt.expected)
		}
	}
}

func TestSortedRecordSet(t *testing.T) {
	var keySeq uint64 = 0

	numKeys := make([]int, 1024)
	for i := range numKeys {
		numKeys[i] = i + 1
	}

	for _, numKey := range numKeys {
		var sortedRecords []*Record
		var sortedValues []string
		for r := 0; r < numKey; r++ {
			sortedRecords = append(sortedRecords, &Record{
				Key:   []byte(utils.GenerateSeqString(&keySeq, 16, 5)),
				Value: []byte(fmt.Sprint(r)),
			})

			sortedValues = append(sortedValues, fmt.Sprint(r))
		}

		_, tree := dummyTree(t, "TestSortedLoad", 4)

		err := tree.SortedRecordSet(sortedRecords)
		if err != nil {
			t.Fatal(err)
		}

		if err := tree.TreeVerify(); err != nil {
			t.Fatal(err)
		}

		startKey := sortedRecords[0].Key
		endKey := sortedRecords[len(sortedRecords)-1].Key

		rangeScan := tree.RangeScan([]byte(startKey), []byte(endKey))
		if len(sortedValues) != len(rangeScan) {
			t.Fatal("Range should be equal", len(sortedValues), len(rangeScan))
		}

		for i, s := range sortedRecords {
			if string(rangeScan[i].Value) != string(s.Value) || string(rangeScan[i].Key) != string(s.Key) {
				t.Fatal("Record should be equal")
			}
		}
	}
}

func TestDelete(t *testing.T) {
	var keySeq uint64 = 0
	var sortedRecords []*Record
	var sortedKeys [][]byte
	var sortedValues []string
	for r := 0; r < 5120; r++ {
		key := []byte(utils.GenerateSeqString(&keySeq, 16, 5))
		sortedKeys = append(sortedKeys, key)

		sortedRecords = append(sortedRecords, &Record{
			Key:   key,
			Value: key,
		})

		sortedValues = append(sortedValues, fmt.Sprint(r))
	}

	var shuffledKeys [][][]byte
	for i := 0; i < 10; i++ {
		shuffledKeys = append(shuffledKeys, utils.Shuffle(sortedKeys)[:rand.IntN(len(sortedKeys))])
	}

	for i, keys := range shuffledKeys {

		t.Log(utils.ArrayToStrings(keys))

		_, tree := dummyTree(t, "TestDelete", 8)

		if i%2 == 1 {
			for _, r := range sortedRecords {
				tree.Set(r.Key, r.Value)
			}
		} else {
			tree.SortedRecordSet(sortedRecords)
		}

		for _, k := range keys {
			rec, err := tree.Get(k)
			if err != nil || bytes.Compare(rec.Value, k) != 0 {
				t.Fatal(err)
			}
			err = tree.Delete(k)
			if err != nil {
				t.Fatal(err)
			}
			_, err = tree.Get(k)
			if err == nil {
				t.Fatal()
			}

			if err := tree.TreeVerify(); err != nil {
				t.Fatal(err)
			}
		}
	}

	// for _, keys := range shuffledKeys {

	// 	t.Log(utils.ArrayToStrings(keys))

	// 	_, tree := dummyTree(t, "TestDelete", 4)

	// 	tree.SortedRecordSet(sortedRecords)

	// 	for _, k := range keys {
	// 		err := tree.Delete(k)
	// 		if err != nil {
	// 			t.Fatal(err, utils.ArrayToStrings(keys))
	// 		}
	// 	}

	// 	if err := tree.TreeVerify(); err != nil {
	// 		t.Fatal(err)
	// 	}
	// }
}

func TestSplitInternal(t *testing.T) {
	_, tree := dummyTree(t, "TestSplitInternal", 4)

	var keySeq uint64 = 0
	var sortedRecords []*Record
	var sortedKeys [][]byte
	var sortedValues []string
	for r := 0; r < 64; r++ {

		key := []byte(utils.GenerateSeqString(&keySeq, 16, 5))
		sortedKeys = append(sortedKeys, key)

		sortedRecords = append(sortedRecords, &Record{
			Key:   key,
			Value: []byte(fmt.Sprint(r + 1)),
		})

		sortedValues = append(sortedValues, fmt.Sprint(r))
	}

	for _, r := range sortedRecords {
		tree.Set(r.Key, r.Value)
	}

	tree.Set([]byte("0000000000000196"), []byte("Hello:196"))
	tree.Set([]byte("0000000000000197"), []byte("Hello:197"))
	tree.Set([]byte("0000000000000198"), []byte("Hello:198"))
	tree.Set([]byte("0000000000000199"), []byte("Hello:199"))
	if err := tree.TreeVerify(); err != nil {
		t.Fatal(err)
	}

	nodes := tree.GetFirstNodePerHeight()
	expected := []uint64{21, 7, 2, 0}
	if len(nodes) != len(expected) {
		t.Fatalf("Expected %d nodes, got %d", len(expected), len(nodes))
	}
	for i, node := range nodes {
		if node.NodeID != expected[i] {
			t.Errorf("At height %d, expected NodeID %d, got %d", i, expected[i], node.NodeID)
		}
	}
}
