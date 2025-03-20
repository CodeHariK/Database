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
	batchNumLevel uint8,
	batchBaseSize uint32,
	batchIncrement uint8,
	batchLength uint8,
) (*BTree, error) {
	if order < MIN_ORDER || order > MAX_ORDER {
		return nil, ErrorInvalidOrder
	}

	if batchIncrement < 110 || batchIncrement > 200 {
		return nil, ErrorInvalidBatchIncrement
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

		Order:          order,
		BatchNumLevel:  batchNumLevel,
		BatchBaseSize:  batchBaseSize,
		BatchIncrement: batchIncrement,
		BatchLength:    batchLength,

		nodeSize: uint32(nodeSize),

		minNumKeys: uint32(int(order)-1) / 2,
	}

	nodePager, err := tree.NewNodePager("index", 0)
	if err != nil {
		return nil, err
	}
	tree.nodePager = nodePager

	recordPagers := make([]*RecordPager, batchNumLevel)
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

func (tree *BTree) createHeader() ([]byte, error) {
	header64, err := binstruct.Serialize(*tree)
	if err != nil {
		return nil, err
	}

	header64 = append([]byte(SECRETARY), header64...)

	if len(header64) < SECRETARY_HEADER_LENGTH {
		// header = append(header, make([]byte, rootHeaderSize-len(header))...)
		header64 = append(header64, utils.MakeByteArray(SECRETARY_HEADER_LENGTH-len(header64), '-')...)
	}

	return header64, nil
}

func (tree *BTree) SaveHeader() error {
	header, err := tree.createHeader()
	if err != nil {
		return err
	}

	return tree.nodePager.WriteAt(header, 0)
}

func (tree *BTree) readRoot() error {
	rootBytes, err := tree.nodePager.ReadAt(SECRETARY_HEADER_LENGTH, int32(tree.nodeSize))
	if err != nil {
		return err
	}

	var root Node
	err = binstruct.Deserialize(rootBytes, &root)
	if err != nil {
		return err
	}

	tree.root = &root
	return nil
}

func (tree *BTree) saveRoot() error {
	rootHeader, err := binstruct.Serialize(*tree.root)
	if err != nil {
		return err
	}

	return tree.nodePager.WriteAt(rootHeader, SECRETARY_HEADER_LENGTH)
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

func (t *BTree) Height() int {
	if t.root == nil {
		return 0
	}

	height := 0
	node := t.root
	for node != nil {
		height++
		if len(node.children) == 0 { // Leaf node reached
			break
		}
		node = node.children[0] // Move to the first child
	}

	return height
}
