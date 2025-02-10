package secretary

import (
	"fmt"
	"log"
	"os"

	"github.com/codeharik/secretary/utils"
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

			tree, err := NewBTreeReadHeader(file.Name())
			if err == nil && tree.CollectionName == file.Name() {
				secretary.trees[file.Name()] = tree
				fmt.Print("\n[DIR] * ", file.Name())

			} else {
				fmt.Print("\n[DIR] ", file.Name(), " ", err)
			}
		}
	}

	users := secretary.trees["users"]
	key1 := []byte(utils.GenerateRandomString(16))
	value1 := []byte("Hello world!")
	err = users.Insert(key1, value1)
	if err != nil {
		fmt.Print(err)
	}

	return secretary
}
