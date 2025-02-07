package secretary

import (
	"fmt"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/binstruct"
)

func NewBTree(
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
		return nil, ErrorInvalidCollectionName()
	}

	if err := utils.EnsureDir(fmt.Sprintf("%s/%s", SECRETARY, safeCollectionName)); err != nil {
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
	}

	nodeBatchStore, err := tree.NewBatchStore("index", 0)
	if err != nil {
		return nil, err
	}

	tree.nodeBatchStore = nodeBatchStore

	recordBatchStores := make([]*BatchStore, batchNumLevel)
	for i := range recordBatchStores {
		store, err := tree.NewBatchStore("record", uint8(i))
		if err != nil {
			return nil, err
		}

		recordBatchStores[i] = store
	}
	tree.recordBatchStores = recordBatchStores

	return tree, nil
}

func (tree *BTree) close() error {
	if err := tree.nodeBatchStore.file.Close(); err != nil {
		return err
	}

	for _, batchStore := range tree.recordBatchStores {
		if err := batchStore.file.Close(); err != nil {
			return err
		}
	}

	return nil
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

	return tree.nodeBatchStore.WriteAt(0, header)
}
