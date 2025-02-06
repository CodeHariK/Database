package secretary

import (
	"bytes"

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
			return nil, ErrorInvalidKeySize
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

// TODO : Binary search key
func (tree *bTree) searchNode(n *node, key []byte) (*node, error) {
	if len(key) != KEY_SIZE {
		return nil, ErrorInvalidKeySize
	}

	tree.searchKey(tree.root, key)

	return nil, ErrorNodeNotInTree
}

func (tree *bTree) dataLocationCheck(location DataLocation) error {
	if location == -1 {
		return ErrorInvalidDataLocation
	}
	return nil
}

func (tree *bTree) addKey(n *node, key []byte, keyOffset DataLocation) error {
	if (n.NumKeys + 1) > tree.Order {
		return ErrorNumKeysMoreThanOrder
	}
	if len(key) != KEY_SIZE {
		return ErrorInvalidKeySize
	}
	if err := tree.dataLocationCheck(keyOffset); err != nil {
		return err
	}

	n.NumKeys += 1

	n.KeyOffsets = append(n.KeyOffsets, keyOffset)
	n.Keys = append(n.Keys, key)

	return nil
}

// TODO : Binary search key
// Returns equal or less than
func (tree *bTree) searchKey(n *node, key []byte) (int, error) {
	if len(key) != KEY_SIZE {
		return -1, ErrorInvalidKeySize
	}

	for i, k := range n.Keys {
		if bytes.Compare(key, k) == 0 {
			return i, nil
		}
	}

	return -1, ErrorKeyNotInNode
}

// TODO : Remove Key
func (tree *bTree) removeKey(n *node, key []byte) {
}
