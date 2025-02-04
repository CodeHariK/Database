package binstruct

import (
	"bytes"
	"encoding/json"
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

	Fstring_4_30 string `bin:"Fstring_4_30" byte:"4" max:"30"`
	Fstring_30   string `bin:"Fstring_30" max:"30"`
	Fstring_10   string `bin:"Fstring_10" max:"10"`

	Fbytes     []byte `bin:"Fbytes"`
	Fbytes_300 []byte `bin:"Fbytes_300" max:"300"`

	Fstring_Empty string `bin:"FstringEmpty"`
	Fbytes_Empty  []byte `bin:"FbytesEmpty"`

	Fint64_array    []int64 `bin:"Fint64_array"`
	Fint32_array_20 []int32 `bin:"Fint32_array_20" max:"20"`
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
				Fint64_array:    []int64{125, 2000},
				Fint32_array_20: utils.GenerateRandomSlice[int32](20),
			},
		},
		"Array truncate": {
			false,
			testStruct{
				Fint32_array_20: utils.GenerateRandomSlice[int32](21),
			},
		},
	}

	for _, test := range tests {

		t.Logf("\n----------------------------------------\nOriginal: %+v", test.s)

		binaryData, _ := Serialize(test.s)
		t.Logf("\n\nSerializ: %+v\n", binaryData)

		var d testStruct
		Deserialize(binaryData, &d)
		t.Logf("\n\nDeserial: %+v\n", d)

		hashS, _ := utils.Md5Struct(test.s)
		hashD, _ := utils.Md5Struct(d)

		eq, err := Compare(test.s, d)
		if test.equal != eq || err != nil || test.equal != (hashS == hashD) {
			t.Fatalf("\nShould be Equal : %v , %s == %s\n", test.equal, hashS, hashD)
		}
	}
}

func TestCompareBinaryStruct(t *testing.T) {
	tests := map[string]struct {
		equal bool
		a     testStruct
		b     testStruct
	}{
		"Equal": {
			true,
			testStruct{
				Fint8:           -5,
				Fuint8:          200,
				Fint16:          -3000,
				Fuint16:         60000,
				Fint32:          -2000000000,
				Fuint32:         4000000000,
				Fint64:          -9000000000000000000,
				Fuint64:         18000000000000000000,
				Ffloat64:        3.1415926535,
				Fstring:         "Hello",
				Fstring_4_30:    "Hello",
				Fstring_30:      "Hello",
				Fbytes:          []byte{0x12, 0x34, 0x56, 0x78},
				Fint64_array:    []int64{125, 2000},
				Fint32_array_20: []int32{125, 2000},
			},
			testStruct{
				Fint8:           -5,
				Fuint8:          200,
				Fint16:          -3000,
				Fuint16:         60000,
				Fint32:          -2000000000,
				Fuint32:         4000000000,
				Fint64:          -9000000000000000000,
				Fuint64:         18000000000000000000,
				Ffloat64:        3.1415926535,
				Fstring:         "Hello",
				Fstring_4_30:    "Hello",
				Fstring_30:      "Hello",
				Fbytes:          []byte{0x12, 0x34, 0x56, 0x78},
				Fint64_array:    []int64{125, 2000},
				Fint32_array_20: []int32{125, 2000},
			},
		},
		"NotEqual": {
			false,
			testStruct{
				Fbytes_300: make([]byte, 300),
			},
			testStruct{
				Fint8: -5,
			},
		},
	}

	for _, test := range tests {

		t.Logf("\n----------------------------------------\nA: %+v\nB: %+v", test.a, test.b)

		equal, err := Compare(test.a, test.b)
		if err != nil {
			t.Fatal(err)
		}

		if equal != test.equal {
			t.Fatal("Compare failed")
		}
	}
}

func TestMarshalJSONBinaryStruct(t *testing.T) {
	h := testStruct{
		Fint8:           -5,
		Fuint8:          200,
		Fint16:          -3000,
		Fuint16:         60000,
		Fint32:          -2000000000,
		Fuint32:         4000000000,
		Fint64:          -9000000000000000000,
		Fuint64:         18000000000000000000,
		Ffloat64:        3.1415926535,
		Fstring:         "Hello",
		Fstring_4_30:    "Hello",
		Fstring_30:      "Hello",
		Fbytes:          []byte{0x12, 0x34, 0x56, 0x78},
		Fint64_array:    []int64{125, 2000},
		Fint32_array_20: []int32{125, 2000},
	}

	t.Logf("\n----------------------------------------\nOriginal: %+v", h)

	binaryData, _ := Serialize(h)
	t.Logf("\n\nSerializ: %+v\n", binaryData)

	var d testStruct
	Deserialize(binaryData, &d)
	t.Logf("\n\nDeserial: %+v\n", d)

	jsonH, err := MarshalJSON(h)
	if err != nil {
		t.Fatal(err)
	}
	jsonD, err := MarshalJSON(d)
	if err != nil {
		t.Fatal(err)
	}

	serialized, _ := json.Marshal(d)
	t.Log("\n\n", string(serialized), "\n\n")

	if bytes.Compare(jsonH, jsonD) != 0 {
		t.Fatalf("Compare failed, %s \n %s", string(jsonH), string(jsonD))
	}
}
