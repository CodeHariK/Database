package secretary

import (
	"bytes"
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

	multipleKeys := make([][]byte, 40)
	for i := range multipleKeys {
		multipleKeys[i] = []byte(utils.GenerateSeqRandomString(16, 4))
		err = users.Insert(multipleKeys[i], multipleKeys[i])
		if err != nil {
			fmt.Printf("Insert failed: %s", err)
		}
	}
	for i := range multipleKeys {
		r, err := users.Search(multipleKeys[i])
		if err != nil || bytes.Compare(r.Value, multipleKeys[i]) != 0 {
			fmt.Printf("\nSearch failed: %d : %s", i, err)
		}
	}

	return secretary
}
