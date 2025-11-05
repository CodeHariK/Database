package secretary

import (
	"fmt"

	"github.com/codeharik/secretary/utils"
)

func SampleSortedKeyRecords(numkeys int) (records []*Record) {
	var keySeq uint64 = 0
	var sortedRecords []*Record
	var sortedValues []string

	for r := 0; r < numkeys; r++ {
		key := []byte(utils.GenerateSeqString(&keySeq, KEY_SIZE, KEY_INCREMENT))

		sortedRecords = append(sortedRecords, &Record{
			Key:   key,
			Value: key,
		})

		sortedValues = append(sortedValues, fmt.Sprint(r))
	}
	return sortedRecords
}

func DummyInitTrees(s *Secretary) {
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

	sortedRecords := SampleSortedKeyRecords(64)
	images.SortedRecordSet(sortedRecords)

	for _, r := range sortedRecords {
		users.SetKV(r.Key, r.Value)
	}

	users.SetKV([]byte("0000000000000196"), []byte("Hello:196"))
	users.SetKV([]byte("0000000000000197"), []byte("Hello:197"))
	users.SetKV([]byte("0000000000000198"), []byte("Hello:198"))
	users.SetKV([]byte("0000000000000199"), []byte("Hello:199"))
}
