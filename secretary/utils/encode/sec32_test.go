package encode

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codeharik/secretary/utils"
)

// Test basic character mappings
func TestEncodingDecoding32(t *testing.T) {
	tests := []struct {
		input string
		sec32 string
		asc32 string
	}{
		{"", "", ""},
		{"| +\n_", "-_p__", "~_p__"},
		{
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			"apcdefghigclmnopcrstuuysyg",
			"apcdefghigclmnopcrstuuysyg",
		},
		{
			"0123456789",
			"0123456789",
			"0123456789",
		},
		{
			`=+-*/\%^<>!?@#$&(),;:'"_. `,
			`epnmd-hmQRiifffmQR________`,
			`epnmd~hm()iifffm()________`,
		},
		{
			"abcdefghijklmnopqrstuvwxyz",
			"apcdefghigclmnopcrstuuysyg",
			"apcdefghigclmnopcrstuuysyg",
		},
		{
			"~`|",
			"-_-",
			"~_~",
		},
	}

	for _, tt := range tests {

		stringToSec32 := StringToSec32(tt.input)
		if stringToSec32 != tt.sec32 {
			utils.Log(t, "inp %q", tt.input, "sec %q", stringToSec32, "exp %q", tt.sec32)
		}

		stringToIndex32 := StringToIndex32(tt.input)
		sec32ToIndex32 := Sec32ToIndex32(stringToSec32)
		if bytes.Compare(stringToIndex32, sec32ToIndex32) != 0 {
			utils.Log(t, "inp %q", tt.input, "sec %q", sec32ToIndex32, "exp %q", stringToIndex32)
		}

		index32ToSec32 := Index32ToSec32(sec32ToIndex32)
		if index32ToSec32 != stringToSec32 {
			utils.Log(t, "sec %q", index32ToSec32, "sec %q", stringToSec32)
		}

		sec32ToAscii32 := Sec32ToAscii32(stringToSec32)
		index32ToAscii32 := Index32ToAscii32(stringToIndex32)
		if sec32ToAscii32 != tt.asc32 || index32ToAscii32 != tt.asc32 {
			utils.Log(t, "inp %q", tt.input, "exp %q", tt.asc32, "asc %q", sec32ToAscii32, "asc %q", index32ToAscii32)
		}

		stringToIndex32Packed := StringToIndex32Packed(tt.input)
		sec32ToIndex32Packed := Sec32ToIndex32Packed(stringToSec32)
		if bytes.Compare(stringToIndex32Packed, sec32ToIndex32Packed) != 0 {
			t.Fatal("Should be equal", stringToIndex32Packed, sec32ToIndex32Packed)
		}

		index32PackedToAscii32 := Index32PackedToAscii32(stringToIndex32Packed)
		index32PackedToSec32 := Index32PackedToSec32(stringToIndex32Packed)
		if !strings.HasPrefix(index32PackedToAscii32, sec32ToAscii32) ||
			!strings.HasPrefix(index32PackedToSec32, stringToSec32) {
			t.Fatal()
		}

		expandStringToSec32 := ExpandStringToSec32(tt.input)
		sec32ToExpandString := Sec32ToExpandString(expandStringToSec32)
		if tt.input != sec32ToExpandString {
			t.Fatal("Should be equal", tt.input, sec32ToExpandString)
		}
	}
}

func TestPackUnpack8to5(t *testing.T) {
	for i := 1; i < 9; i++ {
		arr := utils.GenerateRandomSliceMinMax[byte](i, 0, 127)

		packed := Pack8to5(arr)
		unpacked := Unpack5to8(packed)

		if binStr(unpacked, false, 5) != binStr(packed, false, 8) {
			utils.Log(
				"arr", arr,
				"arr", binStr(arr, true, 8), "",
				"arr", binStr(arr, true, 5), "",
				"pac", binStr(packed, false, 8), "",
				"unp", binStr(unpacked, false, 5),
			)
			t.Fatal()
		}
	}
}
