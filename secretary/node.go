package secretary

import "github.com/codeharik/secretary/utils/binstruct"

func NewNode() *node {
	return &node{}
}

func (tree *bTree) saveRoot() error {
	rootHeader, err := binstruct.Serialize(*tree.root)
	if err != nil {
		return err
	}

	return tree.nodeBatchStore.WriteAt(SECRETARY_HEADER_LENGTH, rootHeader)
}
