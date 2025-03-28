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
