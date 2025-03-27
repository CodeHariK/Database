package secretary

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/binstruct"
	"github.com/codeharik/secretary/utils/file"
)

func (s *Secretary) NewBTree(
	collectionName string,
	order uint8,
	numLevel uint8,
	batchBaseSize uint32,
	increment uint8,
	compactionBatchSize uint32,
) (*BTree, error) {
	if order < MIN_ORDER || order > MAX_ORDER {
		return nil, ErrorInvalidOrder
	}

	if increment < 110 || increment > 200 {
		return nil, ErrorInvalidIncrement
	}

	nodeSize := uint32(order)*(KEY_SIZE+KEY_OFFSET_SIZE) + 3*POINTER_SIZE + 1

	safeCollectionName := utils.SafeCollectionString(collectionName)
	if len(safeCollectionName) < 5 || len(safeCollectionName) > MAX_COLLECTION_NAME_LENGTH {
		return nil, ErrorInvalidCollectionName
	}

	if err := file.EnsureDir(fmt.Sprintf("%s/%s", SECRETARY, safeCollectionName)); err != nil {
		return nil, err
	}

	tree := &BTree{
		CollectionName: safeCollectionName,

		root: &Node{},

		Order:         order,
		NumLevel:      numLevel,
		BatchBaseSize: batchBaseSize,
		Increment:     increment,

		nodeSize: uint32(nodeSize),

		minNumKeys: uint32(int(order)-1) / 2,

		CompactionBatchSize: compactionBatchSize,
	}

	nodePager, err := tree.NewNodePager("index", 0)
	if err != nil {
		return nil, err
	}
	tree.nodePager = nodePager

	recordPagers := make([]*RecordPager, numLevel)
	for i := range recordPagers {
		pager, err := tree.NewRecordPager("record", uint8(i))
		if err != nil {
			return nil, err
		}
		recordPagers[i] = pager
	}
	tree.recordPagers = recordPagers

	return tree, nil
}

func (tree *BTree) close() error {
	errs := []error{}
	if err := tree.nodePager.file.Close(); err != nil {
		errs = append(errs, err)
	}

	for _, pager := range tree.recordPagers {
		if err := pager.file.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (tree *BTree) SaveHeader() error {
	headerBytes, err := binstruct.Serialize(tree)
	if err != nil {
		return err
	}

	headerBytes = append([]byte(SECRETARY), headerBytes...)

	if len(headerBytes) < SECRETARY_HEADER_LENGTH {
		headerBytes = append(headerBytes, utils.MakeByteArray(SECRETARY_HEADER_LENGTH-len(headerBytes), '-')...)
	}

	return tree.nodePager.WriteAt(headerBytes, 0)
}

func (tree *BTree) ReadNodeAtIndex(index uint64) (*Node, error) {
	offset := SECRETARY_HEADER_LENGTH + index*uint64(tree.nodeSize)

	rootBytes, err := tree.nodePager.ReadAt(int64(offset), int32(tree.nodeSize))
	if err != nil {
		return nil, err
	}

	var node Node
	err = binstruct.Deserialize(rootBytes, &node)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func (tree *BTree) readRoot() error {
	node, err := tree.ReadNodeAtIndex(0)
	if err != nil {
		return err
	}
	tree.root = node
	return nil
}

func (tree *BTree) SaveNode(node *Node) error {
	if node.Index == 0 {
		stat, err := tree.nodePager.file.Stat()
		if err != nil {
			return err
		}
		lastFileIndex := uint64(stat.Size()) - uint64(SECRETARY_HEADER_LENGTH)/uint64(tree.nodeSize)
		if lastFileIndex != tree.NumNodeSeq {
			return fmt.Errorf("NumNodes dont match %d != %d", lastFileIndex, tree.NumNodeSeq)
		}
		tree.SaveNodeAtIndex(node, lastFileIndex)
	} else {
		tree.SaveNodeAtIndex(node, uint64(node.Index))
	}
	return nil
}

func (tree *BTree) SaveNodeAtIndex(node *Node, index uint64) error {
	node.Index = index
	if node.parent != nil {
		node.ParentIndex = node.parent.Index
	}
	if node.next != nil {
		node.NextIndex = node.next.Index
	}
	if node.prev != nil {
		node.PrevIndex = node.prev.Index
	}

	rootHeader, err := binstruct.Serialize(node)
	if err != nil {
		return err
	}

	offset := SECRETARY_HEADER_LENGTH + index*uint64(tree.nodeSize)

	return tree.nodePager.WriteAt(rootHeader, int64(offset))
}

func (tree *BTree) saveRoot() error {
	return tree.SaveNodeAtIndex(tree.root, 0)
}

func (s *Secretary) NewBTreeReadHeader(collectionName string) (*BTree, error) {
	temp, err := s.NewBTree(collectionName,
		10,
		0,
		0,
		125,
		0,
	)
	if err != nil {
		return nil, err
	}

	headerData, err := temp.nodePager.ReadAt(0, SECRETARY_HEADER_LENGTH)
	if err != nil {
		return nil, err
	}

	data := bytes.Trim(headerData, "-")[len(SECRETARY):]
	var deserializedTree BTree
	err = binstruct.Deserialize(data, &deserializedTree)
	if err != nil {
		return nil, err
	}

	return &deserializedTree, nil
}

func (tree *BTree) Height() int {
	height := 0

	for node := tree.root; node != nil; node = node.children[0] {
		height++

		// Stop if we've reached a leaf node (no children)
		if len(node.children) == 0 {
			break
		}
	}

	return height
}

func (tree *BTree) Level(node *Node) int {
	if node == nil || tree.root == nil {
		return -1 // Return -1 for invalid nodes
	}

	level := 0
	for node != nil && node != tree.root {
		level++
		node = node.parent
	}

	return level
}

func (tree *BTree) GetFirstNodePerHeight() []*Node {
	var firstNodePerHeight []*Node

	for node := tree.root; node != nil; node = node.children[0] {
		firstNodePerHeight = append(firstNodePerHeight, node)

		// Stop if we've reached a leaf node (no children)
		if len(node.children) == 0 {
			break
		}
	}

	return firstNodePerHeight
}

func (tree *BTree) BFSCompactBatchTraversal() []*Node {
	var compactBatch []*Node

	firstNodePerHeight := tree.GetFirstNodePerHeight()

	if tree.root == nil {
		return compactBatch
	}
	if tree.nextCompactionNode == nil {
		tree.nextCompactionNode = tree.root
	}

	for i := 0; i < int(tree.CompactionBatchSize); i++ {

		// Ensure nextCompactionNode is not nil
		if tree.nextCompactionNode == nil {
			// utils.Log("tree.nextCompactionNode == nil")
			break
		}

		compactBatch = append(compactBatch, tree.nextCompactionNode)

		if tree.nextCompactionNode.next != nil {
			tree.nextCompactionNode = tree.nextCompactionNode.next
		} else {
			// Ensure nextCompactionNode is not nil before calling Level()
			level := tree.Level(tree.nextCompactionNode)

			// utils.Log("level", level, "lenfirst", len(firstNodePerHeight), level > 0 && (level+1) < len(firstNodePerHeight))

			if level >= 0 && (level+1) < len(firstNodePerHeight) {
				firstNode := firstNodePerHeight[level+1]
				tree.nextCompactionNode = firstNode
			} else {
				break
			}
		}
	}

	return compactBatch
}
