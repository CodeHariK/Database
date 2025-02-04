package secretary

import (
	"fmt"

	"github.com/codeharik/secretary/utils"
)

func NewBTree(
	collectionName string,
	order uint8,
	keySize uint8,
	batchNumLevel uint8,
	batchBaseSize uint32,
	batchIncrement uint8,
	batchLength uint8,
) (*bTree, error) {
	if order < MIN_ORDER || order > MAX_ORDER {
		return nil, fmt.Errorf("Order must be between %d and %d", MIN_ORDER, MAX_ORDER)
	}

	if batchIncrement < 110 || batchIncrement > 200 {
		return nil, fmt.Errorf("Batch Increment must be between 110 and 200")
	}
	if keySize != 8 && keySize != 16 {
		return nil, fmt.Errorf("Key Size must be 8 or 16")
	}

	safeCollectionName := utils.SafeCollectionString(collectionName)
	if len(safeCollectionName) < 5 || len(safeCollectionName) > MAX_COLLECTION_NAME_LENGTH {
		return nil, fmt.Errorf("Collection name is not valid, should be a-z 0-9 and with >4 & <100 characters")
	}

	if dirNotExist := utils.EnsureDir(fmt.Sprintf("%s/%s", SECRETARY, safeCollectionName)); dirNotExist != nil {
		return nil, fmt.Errorf(safeCollectionName, dirNotExist.Error())
	}

	tree := &bTree{
		CollectionName: safeCollectionName,

		root: nil,

		Order:          order,
		KeySize:        keySize,
		BatchNumLevel:  batchNumLevel,
		BatchBaseSize:  batchBaseSize,
		BatchIncrement: batchIncrement,
		BatchLength:    batchLength,
	}

	nodeBatchStore, err := tree.NewBatchStore("index", 0)
	if err != nil {
		return nil, err
	}

	tree.nodeBatchStore = nodeBatchStore

	recordBatchStores := make([]*batchStore, batchNumLevel)
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

func (tree *bTree) createHeader() ([]byte, error) {
	header, err := utils.BinaryStructSerialize(*tree)
	if err != nil {
		return nil, err
	}

	header = append([]byte(SECRETARY), header...)

	if len(header) < SECRETARY_HEADER_LENGTH {
		header = append(header, make([]byte, SECRETARY_HEADER_LENGTH-len(header))...)
	}

	return header, nil
}

func (tree *bTree) SaveHeader() error {
	header, err := tree.createHeader()
	if err != nil {
		return err
	}

	return tree.nodeBatchStore.WriteAt(0, header)
}
