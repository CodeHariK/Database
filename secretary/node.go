package secretary

import (
	"github.com/codeharik/secretary/utils/binstruct"
)

func (tree *bTree) NewNode(
	parentOffset, nextOffset, prevOffset DataLocation,
	numKeys int,
	keyOffsets []DataLocation,
	keys [][]byte,
) (*node, error) {
	if numKeys > int(tree.Order) {
		return nil, ErrorNumKeysMoreThanOrder
	}
	if len(keyOffsets) != numKeys || len(keys) != numKeys {
		return nil, ErrorNumKeysNotMatching
	}
	for i, el := range keys {
		if len(el) != KEY_SIZE {
			return nil, ErrorIncorrectKeySize
		}

		if err := tree.dataLocationCheck(keyOffsets[i]); err != nil {
			return nil, err
		}
	}

	return &node{
		ParentOffset: parentOffset,
		NextOffset:   nextOffset,
		PrevOffset:   prevOffset,

		NumKeys:    uint8(numKeys),
		KeyOffsets: keyOffsets,
		Keys:       keys,
	}, nil
}

func (tree *bTree) saveRoot() error {
	rootHeader, err := binstruct.Serialize(*tree.root)
	if err != nil {
		return err
	}

	return tree.nodeBatchStore.WriteAt(SECRETARY_HEADER_LENGTH, rootHeader)
}

func (tree *bTree) dataLocationCheck(location DataLocation) error {
	if location == -1 {
		return ErrorInvalidDataLocation
	}
	return nil
}

func (tree *bTree) addKey(n *node, keyOffset DataLocation, key []byte) error {
	if (n.NumKeys + 1) > tree.Order {
		return ErrorNumKeysMoreThanOrder
	}
	if len(key) != KEY_SIZE {
		return ErrorIncorrectKeySize
	}
	if err := tree.dataLocationCheck(keyOffset); err != nil {
		return err
	}

	n.NumKeys += 1

	n.KeyOffsets = append(n.KeyOffsets, keyOffset)
	n.Keys = append(n.Keys, key)

	return nil
}
