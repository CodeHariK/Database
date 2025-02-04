package secretary

import (
	"bytes"
	"testing"

	"github.com/codeharik/secretary/utils"
)

func TestNodeSerilization(t *testing.T) {
	n := node{
		ParentOffset: 101,
		NextOffset:   102,
		PrevOffset:   103,

		NumKeys: 104,

		KeyOffsets: []int64{2, 3, 4, 5, 6},
	}

	s, err := utils.BinaryStructSerialize(n)
	if err != nil {
		t.Fatal(err)
	}

	var d node
	err = utils.BinaryStructDeserialize(s, &d)
	if err != nil {
		t.Fatal(err)
	}

	nJson, _ := utils.BinaryStructMarshalJSON(n)
	dJson, _ := utils.BinaryStructMarshalJSON(d)
	t.Logf("\n%s\n%s", string(nJson), string(dJson))

	eq, err := utils.BinaryStructCompare(n, d)
	if !eq || bytes.Compare(nJson, dJson) != 0 || err != nil {
		t.Fatal(err)
	}
}
