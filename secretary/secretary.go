package secretary

import (
	"errors"
	"fmt"
	"os"

	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/file"

	_ "go.uber.org/automaxprocs"
)

func New() (*Secretary, error) {
	startMessage := "Hello Secretary!"

	secretary := &Secretary{
		trees: map[string]*BTree{},

		quit: make(chan any),
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
				startMessage += "\n*" + file.Name()
			} else if err != nil {
				startMessage += "\n" + file.Name() + " " + err.Error()
			}
		}
	}

	utils.Log(startMessage)

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

func (s *Secretary) Shutdown() {
	s.PagerShutdown()
	s.ServerShutdown()
}

func (s *Secretary) PagerShutdown() error {
	closingErrors := make([]error, len(s.trees))
	i := 0
	for _, ss := range s.trees {
		if err := ss.close(); err != nil {
			closingErrors[i] = fmt.Errorf("Error closing %s : %s", ss.CollectionName, err.Error())
		}
		i++
	}
	return errors.Join(closingErrors...)
}
