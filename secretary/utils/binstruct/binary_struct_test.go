package binstruct

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"

	"github.com/codeharik/secretary/utils"
)

// Define struct with different types
type testStruct struct {
	Fint8    int8    `bin:"Fint8"`
	Fuint8   uint8   `bin:"Fuint8"`
	Fint16   int16   `bin:"Fint16"`
	Fuint16  uint16  `bin:"Fuint16"`
	Fint32   int32   `bin:"Fint32"`
	Fuint32  uint32  `bin:"Fuint32"`
	Fint64   int64   `bin:"Fint64"`
	Fuint64  uint64  `bin:"Fuint64"`
	Ffloat64 float64 `bin:"Ffloat64"`
	Fstring  string  `bin:"Fstring"`

	Fstring_4_30 string `bin:"Fstring_4_30" lenbyte:"1" max:"30"`
	Fstring_30   string `bin:"Fstring_30" max:"30"`
	Fstring_10   string `bin:"Fstring_10" max:"10"`

	Fbytes     []byte `bin:"Fbytes"`
	Fbytes_300 []byte `bin:"Fbytes_300" lenbyte:"1" max:"300"`

	Fstring_Empty string `bin:"FstringEmpty"`
	Fbytes_Empty  []byte `bin:"FbytesEmpty"`

	Fint64_array    []int64 `bin:"Fint64_array"`
	Fint32_array_20 []int32 `bin:"Fint32_array_20" max:"20"`

	Farraybytearray_Empty [][]byte    `bin:"Farraybytearray_Empty"`
	Farraybytearray       [][]byte    `bin:"Farraybytearray"`
	Farraybytearray_4     [][]byte    `bin:"Farraybytearray_4" array_elem_len:"4"`
	Farrayint32array      [][]int32   `bin:"Farrayint32array"`
	Farrayint32array_5    [][]int32   `bin:"Farrayint32array_5" array_elem_len:"5"`
	Farrayfloat64array    [][]float64 `bin:"Farrayfloat64array"`
	Farrayfloat64array_2  [][]float64 `bin:"Farrayfloat64array_2" array_elem_len:"2"`
}

func TestBinaryStructSerialize(t *testing.T) {
	tests := map[string]struct {
		equal bool
		s     testStruct
	}{
		"Test": {
			true,
			testStruct{
				Fint8:        -5,
				Fuint8:       200,
				Fint16:       -3000,
				Fuint16:      60000,
				Fint32:       -2000000000,
				Fuint32:      4000000000,
				Fint64:       -9000000000000000000,
				Fuint64:      18000000000000000000,
				Ffloat64:     3.1415926535,
				Fstring:      "Hello",
				Fstring_4_30: "Hello",
				Fstring_30:   "Hello",
				Fbytes:       []byte{0x12, 0x34, 0x56, 0x78},
			},
		},
		"Byte Overflow truncate": {
			false,
			testStruct{
				Fbytes_300: make([]byte, 300),
			},
		},
		"Max Overflow truncate": {
			false,
			testStruct{
				Fstring_10: "Hello World!",
			},
		},
		"Array": {
			true,
			testStruct{
				Fint64_array:    utils.GenerateRandomSlice[int64](2000),
				Fint32_array_20: utils.GenerateRandomSlice[int32](20),
			},
		},
		"Array truncate": {
			false,
			testStruct{
				Fint32_array_20: utils.GenerateRandomSlice[int32](21),
			},
		},
		"Array Byte Array": {
			true,
			testStruct{
				Farraybytearray_Empty: [][]byte{},
				Farraybytearray:       [][]byte{{5, 4}, {100, 101, 102, 103}},
				Farraybytearray_4:     [][]byte{{11, 12, 13, 14}, {101, 102, 103, 104}},
				Farrayint32array:      [][]int32{{5, 4}, {100, 101, 102, 103}},
				Farrayint32array_5:    [][]int32{{11, 12, 0, 0, 0}, {101, 102, 103, 104, 105}},
				Farrayfloat64array:    [][]float64{{3.14, 2.74, 5.2334}},
				Farrayfloat64array_2:  [][]float64{{3.14, 2.74}},
			},
		},
		"Array Byte Array Extend": {
			false,
			testStruct{
				Farraybytearray_4:    [][]byte{{5, 4}, {100, 101, 102, 103}},
				Farrayint32array_5:   [][]int32{{5, 4}, {100, 101, 102, 103}},
				Farrayfloat64array_2: [][]float64{{3.14}},
			},
		},
		"Array Byte Array Truncate": {
			false,
			testStruct{
				Farraybytearray_4:    [][]byte{{5, 4, 7, 8, 9}, {100, 101, 102, 103}},
				Farrayint32array_5:   [][]int32{{5, 4, 7, 8, 9, 10}, {100, 101, 102, 103}},
				Farrayfloat64array_2: [][]float64{{3.14, 2.74, 5.2334}},
			},
		},
	}

	for _, test := range tests {
		binaryData, err := Serialize(test.s)
		if err != nil {
			t.Fatal(err)
		}

		{
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(test.s)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("gob", len(buf.Bytes()), "bin", len(binaryData), " bin/gob", 100*len(binaryData)/len(buf.Bytes()))
		}

		var d testStruct
		err = Deserialize(binaryData, &d)
		if err != nil {
			t.Fatal(t, err)
		}

		jsonH, err := MarshalJSON(test.s)
		if err != nil {
			t.Fatal(err)
		}

		jsonD, err := MarshalJSON(d)
		if err != nil {
			t.Fatal(err)
		}

		if test.equal != (bytes.Compare(jsonH, jsonD) == 0) {
			utils.Log("Compare Should be equal", test.equal,
				"jsonH", len(jsonH), string(jsonH), "",
				"jsonD", len(jsonD), string(jsonD), "",
			)
			t.Fatal()
		}

		hashS, _ := hash(test.s)
		hashD, _ := hash(d)
		eq, err := Compare(test.s, d)
		if test.equal != eq || err != nil || test.equal != (hashS == hashD) {
			t.Fatalf("\nShould be Equal : %v , %s == %s\n", test.equal, hashS, hashD)
		}
	}
}

type testStructArr struct {
	Fint8  int8   `bin:"Fint8"`
	Fbytes []byte `bin:"Fbytes"`
}

func TestBinaryStructArrSerialize(t *testing.T) {
	tests := map[string]struct {
		equal bool
		s     []testStructArr
	}{
		"Struct Array": {
			true,
			[]testStructArr{
				{
					13,
					[]byte{14, 15, 16, 17},
				},
				{
					34,
					[]byte{35, 36},
				},
			},
		},
	}

	for _, test := range tests {
		binaryData, err := Serialize(test.s)
		if err != nil {
			t.Fatal(err)
		}

		{
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(test.s)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("gob", len(buf.Bytes()), "bin", len(binaryData), " bin/gob", 100*len(binaryData)/len(buf.Bytes()))
		}

		var d []testStructArr
		err = Deserialize(binaryData, &d)
		if err != nil {
			t.Fatal(t, err)
		}

		jsonH, err := MarshalJSON(test.s)
		if err != nil {
			t.Fatal(err)
		}

		jsonD, err := MarshalJSON(d)
		if err != nil {
			t.Fatal(err)
		}

		if test.equal != (bytes.Compare(jsonH, jsonD) == 0) {
			utils.Log("Compare Should be equal", test.equal,
				"jsonH", len(jsonH), string(jsonH), "",
				"jsonD", len(jsonD), string(jsonD), "",
			)
			t.Fatal()
		}

		hashS, _ := hash(test.s)
		hashD, _ := hash(d)
		eq, err := Compare(test.s, d)
		if test.equal != eq || err != nil || test.equal != (hashS == hashD) {
			t.Fatalf("\nShould be Equal : %v , %s == %s\n", test.equal, hashS, hashD)
		}
	}
}

type testStructStruct struct {
	Fint8  int8          `bin:"Fint8"`
	FInt32 []int32       `bin:"FInt32"`
	Stc    testStructArr `bin:"Stc"`
}

func TestBinaryStructStructSerialize(t *testing.T) {
	tests := map[string]struct {
		equal bool
		s     testStructStruct
	}{
		"Struct Struct": {
			true,
			testStructStruct{
				3,
				[]int32{1, 2},
				testStructArr{
					14,
					[]byte{25, 27},
				},
			},
		},
	}

	for _, test := range tests {
		binaryData, err := Serialize(test.s)
		if err != nil {
			t.Fatal(err)
		}

		{
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(test.s)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("gob", len(buf.Bytes()), "bin", len(binaryData), " bin/gob", 100*len(binaryData)/len(buf.Bytes()))
		}

		var d testStructStruct
		err = Deserialize(binaryData, &d)
		if err != nil {
			t.Fatal(t, err)
		}

		jsonH, err := MarshalJSON(test.s)
		if err != nil {
			t.Fatal(err)
		}

		jsonD, err := MarshalJSON(d)
		if err != nil {
			t.Fatal(err)
		}

		if test.equal != (bytes.Compare(jsonH, jsonD) == 0) {
			utils.Log("Compare Should be equal", test.equal,
				"jsonH", len(jsonH), string(jsonH), "",
				"jsonD", len(jsonD), string(jsonD), "",
			)
			t.Fatal()
		}

		hashS, _ := hash(test.s)
		hashD, _ := hash(d)
		eq, err := Compare(test.s, d)
		if test.equal != eq || err != nil || test.equal != (hashS == hashD) {
			t.Fatalf("\nShould be Equal : %v , %s == %s\n", test.equal, hashS, hashD)
		}
	}
}
