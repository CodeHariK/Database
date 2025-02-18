package utils

import (
	"reflect"
	"testing"
)

func TestStringsToArray(t *testing.T) {
	var keySeq uint64 = 0
	var sortedKeys [][]byte

	for r := 0; r < 26; r++ {
		key := []byte(GenerateSeqString(&keySeq, 16))
		sortedKeys = append(sortedKeys, key)
	}

	shuffledKeys := make([][][]byte, 4)
	for i := range shuffledKeys {
		shuffledKeys[i] = Shuffle(sortedKeys[:5])
	}

	for _, keys := range shuffledKeys {
		srr := ArrayToStrings(keys)
		t.Log(srr)

		arr := StringsToArray[[]byte](srr)
		t.Log(ArrayToStrings(arr))

		if !reflect.DeepEqual(keys, arr) {
			t.Fatal("Should be equal")
		}
	}
}
