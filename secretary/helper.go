package secretary

import (
	"fmt"

	"github.com/codeharik/secretary/utils"
)

func SampleSortedKeyRecords() (keys [][]byte, records []*Record) {
	var keySeq uint64 = 0
	var sortedRecords []*Record
	var sortedKeys [][]byte
	var sortedValues []string

	for r := 0; r < 64; r++ {
		key := []byte(utils.GenerateSeqString(&keySeq, 16, 5))
		sortedKeys = append(sortedKeys, key)

		sortedRecords = append(sortedRecords, &Record{
			Key:   key,
			Value: key,
		})

		sortedValues = append(sortedValues, fmt.Sprint(r))
	}
	return sortedKeys, sortedRecords
}
