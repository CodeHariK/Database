package main

import (
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

	images, err := s.Tree("images")
	if err != nil {
		log.Fatal(err)
	}

	var sortedRecords []*secretary.Record
	var sortedKeys [][]byte
	var sortedValues []string
	for r := 0; r < 26; r++ {

		key := []byte(utils.GenerateSeqString(16))
		sortedKeys = append(sortedKeys, key)

		sortedRecords = append(sortedRecords, &secretary.Record{
			Key:   key,
			Value: []byte(fmt.Sprint(r + 1)),
		})

		sortedValues = append(sortedValues, fmt.Sprint(r))
	}
	images.SortedRecordSet(sortedRecords)

	for _, r := range sortedRecords {
		users.Set(r.Key, r.Value)
	}

	for r := 0; r < 4; r++ {

		err := users.Delete(sortedKeys[r])
		if err != nil {
			fmt.Println(err)
		}
	}

	s.Serve()
}
