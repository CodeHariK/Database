package encode

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codeharik/secretary/utils"
)

// Test basic character mappings
func TestEncodingDecoding16(t *testing.T) {
	tests := []struct {
		input string
		sec16 string
		asc16 string
	}{
		{"", "", ""},
		{"| +\n_", "-_-__", "-_-__"},
		{
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			"abcdefgaegclnnobcrsdovvsvg",
			"abcdefgaegclnnobcrsdovvsvg",
		},
		{
			"0123456789",
			"0000000000",
			"0000000000",
		},
		{
			`=+-*/\%^<>!?@#$&(),;:'"_. `,
			`------------------________`,
			`------------------________`,
		},
		{
			"abcdefghijklmnopqrstuvwxyz",
			"abcdefgaegclnnobcrsdovvsvg",
			"abcdefgaegclnnobcrsdovvsvg",
		},
		{
			"~`|",
			"-_-",
			"-_-",
		},
	}

	for _, tt := range tests {

		stringToSec16 := StringToSec16(tt.input)
		if stringToSec16 != tt.sec16 {
			utils.Log(t, "inp %q", tt.input, "sec %q", stringToSec16, "exp %q", tt.sec16)
		}

		stringToIndex16 := StringToIndex16(tt.input)
		sec16ToIndex16 := Sec16ToIndex16(stringToSec16)
		if bytes.Compare(stringToIndex16, sec16ToIndex16) != 0 {
			utils.Log(t, "inp %q", tt.input, "sec %q", sec16ToIndex16, "exp %q", stringToIndex16)
		}

		index16ToSec16 := Index16ToSec16(sec16ToIndex16)
		if index16ToSec16 != stringToSec16 {
			utils.Log(t, "sec %q", index16ToSec16, "sec %q", stringToSec16)
		}

		sec16ToAscii16 := Sec16ToAscii16(stringToSec16)
		index16ToAscii16 := Index16ToAscii16(stringToIndex16)
		if sec16ToAscii16 != tt.asc16 || index16ToAscii16 != tt.asc16 {
			utils.Log(t, "inp %q", tt.input, "exp %q", tt.asc16, "asc %q", sec16ToAscii16, "asc %q", index16ToAscii16)
		}

		stringToIndex16Packed := StringToIndex16Packed(tt.input)
		sec16ToIndex16Packed := Sec16ToIndex16Packed(stringToSec16)
		if bytes.Compare(stringToIndex16Packed, sec16ToIndex16Packed) != 0 {
			t.Fatal("Should be equal", stringToIndex16Packed, sec16ToIndex16Packed)
		}

		index16PackedToAscii16 := Index16PackedToAscii16(stringToIndex16Packed)
		index16PackedToSec16 := Index16PackedToSec16(stringToIndex16Packed)
		if !strings.HasPrefix(index16PackedToAscii16, sec16ToAscii16) ||
			!strings.HasPrefix(index16PackedToSec16, stringToSec16) {
			t.Fatal()
		}

		expandStringToSec16 := ExpandStringToSec16(tt.input)
		sec16ToExpandString := Sec16ToExpandString(expandStringToSec16)
		if tt.input != sec16ToExpandString {
			t.Fatal("Should be equal", tt.input, sec16ToExpandString)
		}
	}
}

func TestPackUnpack8to4(t *testing.T) {
	for i := 1; i < 9; i++ {
		arr := utils.GenerateRandomSliceMinMax[byte](i, 0, 127)

		packed := Pack8to4(arr)
		unpacked := Unpack4to8(packed)

		if binStr(unpacked, false, 4) != binStr(packed, false, 8) {
			utils.Log(
				"arr", arr,
				"arr", binStr(arr, true, 8), "",
				"arr", binStr(arr, true, 4), "",
				"pac", binStr(packed, false, 8), "",
				"unp", binStr(unpacked, false, 4),
			)
			t.Fatal()
		}
	}
}
