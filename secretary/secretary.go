package secretary

import (
	"fmt"
	"log"
	"os"
)

func New() *Secretary {
	secretary := &Secretary{
		trees: map[string]*BTree{},
	}

	fmt.Println("Hello Secretary!")

	dirPath := "./SECRETARY"

	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			fmt.Print("\n[DIR] ", file.Name())

			tree, err := NewBTreeReadHeader(file.Name())
			if err == nil && tree.CollectionName == file.Name() {
				secretary.trees[file.Name()] = tree
			} else {
				fmt.Print(" ", err)
			}
		}
	}

	return secretary
}
