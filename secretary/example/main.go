package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/codeharik/secretary"
	"github.com/codeharik/secretary/utils"
)

func main() {
	buf := make([]byte, 1024*100) // Large buffer to accommodate more stack frames
	n := runtime.Stack(buf, true)
	fmt.Println(string(buf[:n]))

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
	images.SortedRecordSet(sortedRecords)

	for _, r := range sortedRecords {
		users.Set(r.Key, r.Value)
	}

	s.Serve()
}
