package secretary

import (
	"os"

	"github.com/codeharik/secretary/utils"
)

func New() (*Secretary, error) {
	utils.Log("Hello Secretary!")

	secretary := &Secretary{
		trees: map[string]*BTree{},
	}

	dirPath := "./SECRETARY"

	err := utils.EnsureDir(dirPath)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {

			tree, err := secretary.NewBTreeReadHeader(file.Name())
			if err == nil && tree.CollectionName == file.Name() {
				secretary.trees[file.Name()] = tree
				utils.Log("[DIR] *", file.Name())

			} else {
				// utils.Log("[DIR] ", file.Name(), " ", err)
			}
		}
	}

	return secretary, nil
}

func (s *Secretary) Tree(name string) (*BTree, error) {
	tree, ok := s.trees[name]
	if !ok {
		return nil, ErrorTreeNotFound
	}
	return tree, nil
}

func (s *Secretary) AddTree(tree *BTree) {
	s.trees[tree.CollectionName] = tree
}
