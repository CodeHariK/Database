package main

import (
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
		1000,
	)

	images, imagesErr := s.NewBTree(
		"images",
		8,
		32,
		1024*1024,
		125,
		1000,
	)
	if userErr != nil || imagesErr != nil {
		utils.Log(userErr, imagesErr)
	}

	sortedRecords := secretary.SampleSortedKeyRecords(64)
	images.SortedRecordSet(sortedRecords)

	for _, r := range sortedRecords {
		users.SetKV(r.Key, r.Value)
	}

	users.SetKV([]byte("0000000000000196"), []byte("Hello:196"))
	users.SetKV([]byte("0000000000000197"), []byte("Hello:197"))
	users.SetKV([]byte("0000000000000198"), []byte("Hello:198"))
	users.SetKV([]byte("0000000000000199"), []byte("Hello:199"))

	s.Serve()
}
