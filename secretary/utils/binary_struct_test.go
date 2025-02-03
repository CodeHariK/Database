package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

// Define struct with different types
type Hello struct {
	Fint8        int8    `bin:"Fint8"`
	Fuint8       uint8   `bin:"Fuint8"`
	Fint16       int16   `bin:"Fint16"`
	Fuint16      uint16  `bin:"Fuint16"`
	Fint32       int32   `bin:"Fint32"`
	Fuint32      uint32  `bin:"Fuint32"`
	Fint64       int64   `bin:"Fint64"`
	Fuint64      uint64  `bin:"Fuint64"`
	Ffloat64     float64 `bin:"Ffloat64"`
	Fstring      string  `bin:"Fstring"`
	Fstring_4_30 string  `bin:"Fstring_4_30" byte:"4" size:"30"`
	Fbytes       []byte  `bin:"Fbytes"`
}

// Function to compute MD5 hash of a struct
func hashStruct(data interface{}) (string, error) {
	// Serialize the struct to JSON
	serialized, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Compute MD5 hash of the serialized data
	hash := md5.New()
	hash.Write(serialized)

	// Get the hash sum as a byte slice
	hashBytes := hash.Sum(nil)

	// Return the hash as a hexadecimal string
	return hex.EncodeToString(hashBytes), nil
}

func runTest(t *testing.T, hello Hello, equal bool) {
	t.Logf("\n----------------------------------------\nOriginal: %+v", hello)

	binaryData, _ := SerializeBinary(hello)
	t.Logf("\n\nSerialized Binary: %+v\n", binaryData)

	var newHello Hello
	DeserializeBinary(binaryData, &newHello)
	t.Logf("\n\nDeserialized Struct: %+v\n", newHello)

	hashHello, _ := hashStruct(hello)
	hashNewHello, _ := hashStruct(newHello)

	if equal != reflect.DeepEqual(hello, newHello) {
		t.Fatalf("\nShould be Equal : %v , %s == %s\n", equal, hashHello, hashNewHello)
	}
}

func TestBinaryStruct(t *testing.T) {
	helloTests := map[string]struct {
		equal bool
		hello Hello
	}{
		"Test1": {
			true,
			Hello{
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
				Fbytes:       []byte{0x12, 0x34, 0x56, 0x78},
			},
		},
		"Test2": {
			false,
			Hello{
				Fbytes: make([]byte, 300),
			},
		},
	}

	for _, test := range helloTests {
		runTest(t, test.hello, test.equal)
	}
}
