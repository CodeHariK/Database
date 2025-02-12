package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/codeharik/secretary"
	"github.com/codeharik/secretary/utils"
)

func main() {
	s, err := secretary.New()
	if err != nil {
		log.Fatal(err)
	}

	users, err := s.Tree("users")
	if err != nil {
		log.Fatal(err)
	}

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

	s.Serve()
}
