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

	for r := 0; r < 64; r++ {
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

	users.Set([]byte("0000000000000196"), []byte("Hello:196"))
	users.Set([]byte("0000000000000197"), []byte("Hello:197"))
	users.Set([]byte("0000000000000198"), []byte("Hello:198"))
	users.Set([]byte("0000000000000199"), []byte("Hello:199"))

	s.Serve()
}
