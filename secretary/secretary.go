package secretary

import (
	"os"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/file"

	_ "go.uber.org/automaxprocs"
)

func New() (*Secretary, error) {
	utils.Log("Hello Secretary!")

	secretary := &Secretary{
		trees: map[string]*BTree{},
	}

	dirPath := "./SECRETARY"

	err := file.EnsureDir(dirPath)
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
				secretary.AddTree(tree)
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

func (s *Secretary) Close() {
	for _, ss := range s.trees {
		if err := ss.close(); err != nil {
			utils.Log("Error closing", ss.CollectionName, err)
		}
	}
	s.Shutdown()
}
