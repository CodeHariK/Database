package secretary

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/codeharik/secretary/utils"
)

func NewBTree(
	collectionName string, order uint8, keySize uint8,
	batchNumLevel uint8, batchIncrement uint8, batchLength uint8, batchBaseSize uint32,
) (*bTree, error) {
	if order < MIN_ORDER || order > MAX_ORDER {
		return nil, fmt.Errorf("Order must be between %d and %d", MIN_ORDER, MAX_ORDER)
	}

	if batchIncrement < 110 || batchIncrement > 200 {
		return nil, fmt.Errorf("Batch Increment must be between 110 and 200")
	}

	safeCollectionName := utils.SafeCollectionString(collectionName)
	if len(safeCollectionName) < 5 || len(safeCollectionName) > MAX_COLLECTION_NAME_LENGTH {
		return nil, fmt.Errorf("Collection name is not valid, should be a-z 0-9 and with >4 & <100 characters")
	}

	if dirNotExist := utils.EnsureDir(fmt.Sprintf("SECRETARY/%s", safeCollectionName)); dirNotExist != nil {
		return nil, fmt.Errorf(safeCollectionName, dirNotExist.Error())
	}

	tree := &bTree{
		collectionName: safeCollectionName,

		root: nil,

		order:          order,
		keySize:        keySize,
		batchNumLevel:  batchNumLevel,
		batchBaseSize:  batchBaseSize,
		batchIncrement: batchIncrement,
		batchLength:    batchLength,
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

func (tree *bTree) Serialize() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write "SECRETARY" identifier
	identifier := []byte(SECRETARY)
	buf.Write(identifier)

	// Write fields in order
	binary.Write(buf, binary.LittleEndian, tree.order)
	binary.Write(buf, binary.LittleEndian, tree.keySize)
	binary.Write(buf, binary.LittleEndian, tree.batchNumLevel)
	binary.Write(buf, binary.LittleEndian, tree.batchBaseSize)
	binary.Write(buf, binary.LittleEndian, tree.batchIncrement)
	binary.Write(buf, binary.LittleEndian, tree.batchLength)

	// Serialize collectionName (fixed 100 bytes)
	nameBytes := make([]byte, MAX_COLLECTION_NAME_LENGTH)
	copy(nameBytes, tree.collectionName)
	buf.Write(nameBytes)

	// Ensure exactly 128 bytes
	data := buf.Bytes()
	if len(data) > SECRETARY_HEADER_LENGTH {
		return nil, errors.New("serialization exceeded 64 bytes")
	}

	// Pad if necessary
	for len(data) < SECRETARY_HEADER_LENGTH {
		data = append(data, 0)
	}

	return data, nil
}

func DeserializeBTree(data []byte) (*bTree, error) {
	if len(data) != SECRETARY_HEADER_LENGTH {
		return nil, errors.New("invalid data length")
	}

	reader := bytes.NewReader(data)

	// Read identifier
	identifier := make([]byte, 9)
	if _, err := reader.Read(identifier); err != nil {
		return nil, err
	}
	if string(identifier) != SECRETARY {
		return nil, errors.New("invalid identifier")
	}

	tree := &bTree{}

	// Read fields
	binary.Read(reader, binary.LittleEndian, &tree.order)
	binary.Read(reader, binary.LittleEndian, &tree.keySize)
	binary.Read(reader, binary.LittleEndian, &tree.batchNumLevel)
	binary.Read(reader, binary.LittleEndian, &tree.batchBaseSize)
	binary.Read(reader, binary.LittleEndian, &tree.batchIncrement)
	binary.Read(reader, binary.LittleEndian, &tree.batchLength)

	// Read collectionName (fixed 100 bytes, trim nulls)
	nameBytes := make([]byte, MAX_COLLECTION_NAME_LENGTH)
	if _, err := reader.Read(nameBytes); err != nil {
		return nil, err
	}
	tree.collectionName = string(bytes.Trim(nameBytes, "\x00"))

	return NewBTree(
		tree.collectionName,
		tree.order,
		tree.keySize,
		tree.batchNumLevel,
		tree.batchIncrement,
		tree.batchLength,
		tree.batchBaseSize,
	)
}

func (tree *bTree) SaveHeader() error {
	header, err := tree.Serialize()
	if err != nil {
		return err
	}

	return tree.nodeBatchStore.WriteAt(0, header)
}
