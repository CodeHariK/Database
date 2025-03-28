package secretary

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/binstruct"
)

func dummySecretary(t *testing.T) *Secretary {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func dummyTree(t *testing.T, s *Secretary, order uint8) *BTree {
	tree, err := s.NewBTree(
		utils.GenerateRandomString(16),
		order,
		32,
		1024,
		125,
		20,
	)
	if err != nil {
		t.Fatal(err)
	}

	// if err := tree.SaveHeader(); err != nil {
	// 	t.Fatal(collectionName, err)
	// }
	// if err := tree.writeRoot(); err != nil {
	// 	t.Fatal(collectionName, err)
	// }

	return tree
}

func TestNodeSaveRoot(t *testing.T) {
	s := dummySecretary(t)
	tree := dummyTree(t, s, 10)

	root := Node{
		ParentIndex: 101,
		// NextIndex:   102,
		// PrevIndex:   103,

		Keys:        [][]byte{{10, 21, 32, 34, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		KeyLocation: []uint64{2, 3, 4, 5, 6},
	}
	tree.root = &root

	err := tree.writeRoot()
	if err != nil {
		t.Fatal(err)
	}

	err = tree.readRoot()
	if err != nil {
		t.Fatal(err)
	}

	eq, err := binstruct.Compare(tree.root, &root)
	if !eq || err != nil {
		t.Fatal(err)
	}

	s.PagerShutdown()
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

func TestNodeGetLeafNode(t *testing.T) {
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

func TestNodeSet(t *testing.T) {
	s := dummySecretary(t)
	tree := dummyTree(t, s, 10)

	r, err := tree.Get([]byte(utils.GenerateRandomString(16)))
	if err == nil || r != nil {
		t.Error("expected error and got nil", err, r)
	}

	key := []byte(utils.GenerateRandomString(16))
	value := []byte("Hello world!")
	err = tree.SetKV(key, value)
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
	err = tree.SetKV(key, append(value, []byte("world1")...))
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

	s.PagerShutdown()
}

func TestNodeUpdate(t *testing.T) {
	s := dummySecretary(t)
	tree := dummyTree(t, s, 10)

	key := []byte(utils.GenerateRandomString(16))
	value := []byte("Hello world!")

	err := tree.Update(key, value)
	if err == nil {
		t.Error(err)
	}

	err = tree.SetKV(key, value)
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

	s.PagerShutdown()
}

func TestNodeRangeScan(t *testing.T) {
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

func TestNodeSortedRecordSet(t *testing.T) {
	s := dummySecretary(t)
	tree4 := dummyTree(t, s, 8)
	tree8 := dummyTree(t, s, 8)
	tree16 := dummyTree(t, s, 8)

	trees := []*BTree{tree4, tree8, tree16}

	for numKey := 1; numKey < 500; numKey++ {

		sortedRecords := SampleSortedKeyRecords(int(numKey))

		for i := range trees {
			tree := trees[i]
			tree.Erase()

			err := tree.SortedRecordSet(sortedRecords)
			if err != nil {
				t.Fatal(err, numKey, sortedRecords)
			}

			if errs := tree.TreeVerify(); len(errs) != 0 {
				t.Fatal(errs)
			}

			startKey := sortedRecords[0].Key
			endKey := sortedRecords[len(sortedRecords)-1].Key

			rangeScan := tree.RangeScan([]byte(startKey), []byte(endKey))
			if len(sortedRecords) != len(rangeScan) {
				t.Fatal("Range should be equal", len(sortedRecords), len(rangeScan))
			}

			for i, s := range sortedRecords {
				if string(rangeScan[i].Value) != string(s.Value) || string(rangeScan[i].Key) != string(s.Key) {
					t.Fatal("Record should be equal")
				}
			}
		}
	}
	s.PagerShutdown()
}

func TestNodeDelete(t *testing.T) {
	s := dummySecretary(t)
	tree4 := dummyTree(t, s, 4)
	tree8 := dummyTree(t, s, 8)
	tree16 := dummyTree(t, s, 16)

	trees := []*BTree{
		tree4,
		tree8,
		tree16,
	}

	tests := []int{10, 32, 150}

	for _, numKey := range tests {

		sortedRecords := SampleSortedKeyRecords(numKey)

		for order := range trees {

			tree := trees[order]

			var shuffledRecordsArr [][]*Record
			for i := 0; i < 4; i++ {
				shuffledRecordsArr = append(shuffledRecordsArr, utils.Shuffle(sortedRecords))
			}

			// failed := []string{
			// 	 "0000000000000320 0000000000000165 0000000000000020 0000000000000135 0000000000000010 0000000000000100 0000000000000075 0000000000000005 0000000000000080 0000000000000315 0000000000000060 0000000000000240 0000000000000280 0000000000000200 0000000000000310 0000000000000105 0000000000000265 0000000000000140 0000000000000295 0000000000000245 0000000000000155 0000000000000250 0000000000000125 0000000000000035 0000000000000160 0000000000000055 0000000000000065 0000000000000145 0000000000000290 0000000000000070 0000000000000115 0000000000000260 0000000000000300 0000000000000130 0000000000000095 0000000000000050 0000000000000220",
			// }
			// for _, s := range failed {
			// 	b := utils.Map(strings.Split(s, " "),
			// 		func(s string) []byte {
			// 			return []byte(s)
			// 		})
			// 	shuffledKeys = append(shuffledKeys, b)
			// }

			for i, shuffledRecords := range shuffledRecordsArr {

				tree.Erase()

				if i%2 == 1 {
					for _, r := range sortedRecords {
						err := tree.SetKV(r.Key, r.Value)
						if err != nil {
							t.Fatal(err)
						}
					}
				} else {
					err := tree.SortedRecordSet(sortedRecords)
					if err != nil {
						t.Fatal(err)
					}
				}

				for _, record := range shuffledRecords {
					rec, err := tree.Get(record.Key)
					if err != nil || bytes.Compare(rec.Value, record.Key) != 0 {
						t.Fatal(err)
					}

					err = tree.Delete(record.Key)
					if err != nil {
						t.Fatal(err)
					}

					_, err = tree.Get(record.Key)
					if err == nil {
						t.Fatal(err)
					}

					if errs := tree.TreeVerify(); len(errs) != 0 {
						t.Fatal(errs)
					}
				}
			}

		}
	}

	s.PagerShutdown()
}

func TestNodeSplitInternal(t *testing.T) {
	s := dummySecretary(t)
	tree := dummyTree(t, s, 4)

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
		tree.SetKV(r.Key, r.Value)
	}

	tree.SetKV([]byte("0000000000000196"), []byte("Hello:196"))
	tree.SetKV([]byte("0000000000000197"), []byte("Hello:197"))
	tree.SetKV([]byte("0000000000000198"), []byte("Hello:198"))
	tree.SetKV([]byte("0000000000000199"), []byte("Hello:199"))
	if errs := tree.TreeVerify(); len(errs) != 0 {
		t.Fatal(errs)
	}

	{
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

	{ // Perform batch traversal
		compactBatch := tree.BFSCompactBatchTraversal()
		expected := []uint64{21, 7, 20, 34, 47, 2, 6, 11, 15, 19, 25, 29, 50, 33, 38, 42, 46, 0, 1, 3}
		if len(compactBatch) != len(expected) {
			t.Fatalf("Expected %d nodes, got %d", len(expected), len(compactBatch))
		}
		for i, node := range compactBatch {
			if node.NodeID != expected[i] {
				t.Errorf("At index %d, expected NodeID %d, got %d", i, expected[i], node.NodeID)
			}
		}

		compactBatch = tree.BFSCompactBatchTraversal()
		expected = []uint64{4, 5, 8, 9, 10, 12, 13, 14, 16, 17, 18, 22, 23, 24, 26, 27, 28, 48, 49, 30}

		if len(compactBatch) != len(expected) {
			t.Fatalf("Expected %d nodes, got %d", len(expected), len(compactBatch))
		}
		for i, node := range compactBatch {
			if node.NodeID != expected[i] {
				t.Errorf("At index %d, expected NodeID %d, got %d", i, expected[i], node.NodeID)
			}
		}

		compactBatch = tree.BFSCompactBatchTraversal()
		expected = []uint64{31, 32, 35, 36, 37, 39, 40, 41, 43, 44, 45}

		// utils.Log(utils.Map(compactBatch, func(s *Node) uint64 { return s.NodeID }))

		if len(compactBatch) != len(expected) {
			t.Fatalf("Expected %d nodes, got %d", len(expected), len(compactBatch))
		}
		for i, node := range compactBatch {
			if node.NodeID != expected[i] {
				t.Errorf("At index %d, expected NodeID %d, got %d", i, expected[i], node.NodeID)
			}
		}
	}

	s.PagerShutdown()
}
