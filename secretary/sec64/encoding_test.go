package sec64

import (
	"fmt"
	"strings"
	"testing"

	"github.com/codeharik/secretary/utils"
)

// Test basic character mappings
func TestEncodingDecoding(t *testing.T) {
	tests := []struct {
		input  string
		encExp string
		ascExp string
	}{
		{"Kello WORLD!", "kelloSworldK", "kello world!"},
		// {"123\n", "123N", "123\n"},
		// {"[test] {code}", "RtestTSRcodeT", "(test) (code)"},
		// {`=+-*/\%^<>!?@#$&(),;:'"_.`, `ABCDEFGHIJKLMOPQRTUVWXYZ.`, `=+-*/\%^<>!?@#$&(),;:'"_.`},
		// {"_~", "Z_", "_\x00"},
		// {"\x00", "_", "\x00"},
	}

	for _, tt := range tests {
		enc := EncodeString(tt.input)
		if enc != tt.encExp {
			t.Errorf("encode(%q) = %q; want %q", tt.input, enc, tt.encExp)
		}

		dec := DecodeString(enc)
		if dec != tt.ascExp {
			t.Errorf("decode(%q) = %q; want %q", enc, dec, tt.ascExp)
		}

		encoded := Encode(tt.input)

		decoded := Decode(encoded)

		utils.Log(
			"enc", enc,
			"[]byte(enc)", []byte(enc),
			"Pack8to6(enc)", binStr(Pack8to6([]byte(enc)), true, true, false),
			"encoded", encoded,
			"decoded", decoded,
		)

		if decoded != dec {
			t.Errorf("decode(%q) \n %q \n %q want", encoded, decoded, dec)
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
			sb.WriteString(fmt.Sprintf(format+"(%q) ", b, Sec2Ascii[b&0b00111111]))
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

		packed := Pack8to6(test.input)

		unpacked := Unpack6to8(packed)
		if binStr(unpacked, false, false, false) != binStr(packed, true, false, false) {
			t.Fatalf("unpack\n%v\n%v",
				binStr(packed, true, false, false),
				binStr(unpacked, false, false, false),
			)
		}
	}
}
