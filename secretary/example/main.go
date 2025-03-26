package main

import (
	"fmt"
	"os"

	"github.com/codeharik/secretary"
	"github.com/codeharik/secretary/utils"
)

func main() {
	s, err := secretary.New()
	if err != nil {
		utils.Log(err)
		os.Exit(1)
	}

	users, userErr := s.NewBTree(
		"users",
		4,
		32,
		1024,
		125,
		10,
		1000,
	)

	images, imagesErr := s.NewBTree(
		"images",
		4,
		32,
		1024*1024,
		125,
		10,
		1000,
	)
	if userErr != nil || imagesErr != nil {
		utils.Log(userErr, imagesErr)
	}

	s.AddTree(users)
	s.AddTree(images)

	var keySeq uint64 = 0
	var sortedRecords []*secretary.Record
	var sortedKeys [][]byte
	var sortedValues []string

	NumKeys := 10 // 64

	for r := 0; r < NumKeys; r++ {
		key := []byte(utils.GenerateSeqString(&keySeq, 16, 5))
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

	users.Set([]byte("0000000000000016"), []byte("Hello:16"))
	users.Set([]byte("0000000000000017"), []byte("Hello:17"))
	users.Set([]byte("0000000000000018"), []byte("Hello:18"))

	s.Serve()
}
