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
	)

	images, imagesErr := s.NewBTree(
		"images",
		4,
		32,
		1024*1024,
		125,
		10,
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
	for r := 0; r < 26; r++ {

		key := []byte(utils.GenerateSeqString(&keySeq, 16))
		sortedKeys = append(sortedKeys, key)

		sortedRecords = append(sortedRecords, &secretary.Record{
			Key:   key,
			Value: []byte(fmt.Sprint(r + 1)),
		})

		sortedValues = append(sortedValues, fmt.Sprint(r))
	}
	images.SortedRecordSet(sortedRecords)

	// users.SortedRecordSet(sortedRecords)
	for _, r := range sortedRecords {
		users.Set(r.Key, r.Value)
	}

	// for _, k := range utils.Shuffle(sortedKeys[:6]) {
	// 	err := users.Delete(k)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	// for _, k := range utils.Shuffle(sortedKeys[len(sortedKeys)-1:]) {
	// 	err := users.Delete(k)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	// for _, k := range utils.StringsToArray[[]byte](
	// 	[]string{
	// 		"0000000000000026", "0000000000000020", "0000000000000021", "0000000000000018", "0000000000000019",
	// 		"0000000000000022", "0000000000000024", "0000000000000023", "0000000000000025", "0000000000000017",
	// 	},
	// ) {
	// 	err := users.Delete(k)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	s.Serve()
}
