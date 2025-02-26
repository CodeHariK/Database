package encode

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// Test basic character mappings
func TestEncodingDecoding(t *testing.T) {
	tests := []struct {
		input string
		sec64 string
		ascii string
	}{
		// {"", "", ""},
		{"| +\n_0", "F_B+X0", "\\ +\n_0"},
		{
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			"abcdefghijklmnopqrstuvwxyz",
			"abcdefghijklmnopqrstuvwxyz",
		},
		{
			"0123456789",
			"0123456789",
			"0123456789",
		},
		{
			`=+-*/\%^<>!?@#$&(),;:'"_. `,
			`ABCDEFGHIJKLMNOPQRSTUVWX._`,
			`=+-*/\%^<>!?@#$&(),;:'"_. `,
		},
		{
			"abcdefghijklmnopqrstuvwxyz",
			"abcdefghijklmnopqrstuvwxyz",
			"abcdefghijklmnopqrstuvwxyz",
		},
		{
			"~`|",
			"--F",
			"~~\\",
		},
	}

	for _, tt := range tests {
		asciiToSec64 := AsciiToSec64(tt.input)
		asciiToIndex := Ascii64ToIndex(tt.input)
		sec64ToAscii := Sec64ToAscii(asciiToSec64)
		sec64ToIndex := Sec64ToIndex(asciiToSec64)
		// packedIndexes := Pack8to6(asciiToIndex)
		// unpackedIndexes := Unpack6to8(packedIndexes)
		// indexToAscii := IndexToAscii(unpackedIndexes)
		// indexToSec64 := IndexToSec64(unpackedIndexes)
		encoded := Pack8to6(Ascii64ToIndex(tt.input))
		decodedAscii := IndexToAscii64(Unpack6to8(encoded))
		decodedSec64 := IndexToSec64(Unpack6to8(encoded))

		// utils.Log(
		// 	"tt.input      ", tt.input, "",
		// 	"asciiToSec64  ", asciiToSec64, "",
		// 	// "asciiToIndex", asciiToIndex,"",
		// 	"sec64ToAscii  ", sec64ToAscii, "",
		// 	// "sec64ToIndex", sec64ToIndex,"",
		// 	// "packedIndexes", packedIndexes,"",
		// 	// "unpackedIndexes", unpackedIndexes,"",
		// 	// "indexToAscii  ", indexToAscii,"",
		// 	// "indexToSec64  ", indexToSec64,"",
		// 	// "encoded", encoded,"",
		// 	"decodedAscii  ", decodedAscii, "",
		// 	"decodedSec64  ", decodedSec64, "",
		// )

		if bytes.Compare(asciiToIndex, sec64ToIndex) != 0 || // bytes.Compare(asciiToIndex, unpackedIndexes) != 0 ||
			asciiToSec64 != tt.sec64 ||
			sec64ToAscii != tt.ascii ||
			!strings.HasPrefix(decodedAscii, sec64ToAscii) ||
			!strings.HasPrefix(decodedSec64, asciiToSec64) {
			t.Fatal()
		}

	}
}

// binStr formats byte slices into binary strings with different modes
func binStr(data []byte, bytemode, spacemode, debugmode bool) string {
	var sb strings.Builder

	format := "%06b" // Default 6-bit mode
	if bytemode {
		format = "%08b" // 8-bit mode
	}
	if spacemode {
		format += " " // Add space after each formatted byte
	}

	for _, b := range data {
		if debugmode {
			sb.WriteString(fmt.Sprintf(format+"(%q) ", b, ASCII64[SEC64Index[b&0b00111111]]))
		} else {
			sb.WriteString(fmt.Sprintf(format, b))
		}
	}

	return sb.String()
}

func TestPackUnpack(t *testing.T) {
	tests := []struct {
		input []byte
	}{
		{[]byte{}},
		{[]byte{29}},
		{[]byte{53, 2}},
		{[]byte{23, 4, 63}},
		{[]byte{3, 0, 8, 10}},
		{[]byte{30, 41, 0, 8, 10}},
	}

	for _, test := range tests {
		packUnpackCheck(t, test.input)
	}
}

func packUnpackCheck(t *testing.T, input []byte) {
	packed := Pack8to6(input)

	unpacked := Unpack6to8(packed)
	if binStr(unpacked, false, false, false) != binStr(packed, true, false, false) {
		t.Fatalf("unpack\n%v\n%v",
			binStr(packed, true, false, false),
			binStr(unpacked, false, false, false),
		)
	}
}

func TestExpand(t *testing.T) {
	tests := []string{
		"Hello",
		"Hello34",
		"Hello_32",
		"H|llo%32~",
	}
	for _, test := range tests {
		sec := AsciiToSec64Expand(test)

		ascii := Sec64ToAsciiExpand(sec)

		if test != ascii {
			t.Fatal("Should be equal", test, ascii)
		}
	}
}
