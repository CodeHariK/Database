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

	for i := 0; i < 2; i++ {
		key := []byte(utils.GenerateSeqRandomString(16, 4))
		err = users.Insert(key, key)
		if err != nil {
			fmt.Print(err)
		}
	}

	return secretary
}
