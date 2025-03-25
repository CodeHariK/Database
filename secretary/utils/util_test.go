package utils

import (
	"reflect"
	"testing"
)

func TestStringsToArray(t *testing.T) {
	var keySeq uint64 = 0
	var sortedKeys [][]byte

	for r := 0; r < 26; r++ {
		key := []byte(GenerateSeqString(&keySeq, 16, 5))
		sortedKeys = append(sortedKeys, key)
	}

	for i := 0; i < 5; i++ {
		keys := Shuffle(sortedKeys[:5])
		srr := ArrayToStrings(keys)
		arr := StringsToArray[[]byte](srr)
		if !reflect.DeepEqual(keys, arr) {
			t.Fatal("Should be equal")
		}
	}
}
