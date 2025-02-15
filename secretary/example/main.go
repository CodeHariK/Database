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
	var sortedValues []string
	for r := 0; r < 26; r++ {
		sortedRecords = append(sortedRecords, &secretary.Record{
			Key:   []byte(utils.GenerateSeqString(16)),
			Value: []byte(fmt.Sprint(r + 1)),
		})

		sortedValues = append(sortedValues, fmt.Sprint(r))
	}
	images.SortedRecordLoad(sortedRecords)

	for _, r := range sortedRecords {
		users.Insert(r.Key, r.Value)
	}

	s.Serve()
}
