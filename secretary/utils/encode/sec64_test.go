package encode

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/codeharik/secretary/utils"
)

// Test basic character mappings
func TestEncodingDecoding64(t *testing.T) {
	tests := []struct {
		input string
		sec64 string
		asc64 string
	}{
		{"", "", ""},
		// {"| +\n_0", "F_B+X0", "\\ +\n_0"},
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
			"-WF",
			"~\"\\",
		},
	}

	for _, tt := range tests {

		stringToSec64 := StringToSec64(tt.input)
		if stringToSec64 != tt.sec64 {
			utils.Log(t, "inp %q", tt.input, "sec %q", stringToSec64, "exp %q", tt.sec64)
		}

		stringToIndex64 := StringToIndex64(tt.input)
		sec64ToIndex64 := Sec64ToIndex64(stringToSec64)
		if bytes.Compare(stringToIndex64, sec64ToIndex64) != 0 {
			t.Fatal("Should be equal", stringToIndex64, sec64ToIndex64)
		}

		index64ToSec64 := Index64ToSec64(sec64ToIndex64)
		if index64ToSec64 != stringToSec64 {
			t.Fatal("Should be equal", index64ToSec64, stringToSec64)
		}

		sec64ToAscii64 := Sec64ToAscii64(stringToSec64)
		index64ToAscii64 := Index64ToAscii64(stringToIndex64)
		if sec64ToAscii64 != tt.asc64 || index64ToAscii64 != tt.asc64 {
			t.Fatal("Should be equal", tt.asc64, sec64ToAscii64, index64ToAscii64)
		}

		stringToIndex64Packed := StringToIndex64Packed(tt.input)
		sec64ToIndex64Packed := Sec64ToIndex64Packed(stringToSec64)
		if bytes.Compare(stringToIndex64Packed, sec64ToIndex64Packed) != 0 {
			t.Fatal("Should be equal", stringToIndex64Packed, sec64ToIndex64Packed)
		}

		index64PackedToAscii64 := Index64PackedToAscii64(stringToIndex64Packed)
		index64PackedToSec64 := Index64PackedToSec64(stringToIndex64Packed)
		if !strings.HasPrefix(index64PackedToAscii64, sec64ToAscii64) ||
			!strings.HasPrefix(index64PackedToSec64, stringToSec64) {
			t.Fatal()
		}

		expandStringToSec64 := ExpandStringToSec64(tt.input)
		sec64ToExpandString := Sec64ToExpandString(expandStringToSec64)
		if tt.input != sec64ToExpandString {
			t.Fatal("Should be equal", tt.input, sec64ToExpandString)
		}
	}
}

func TestPackUnpack8to6(t *testing.T) {
	for i := 1; i < 400; i++ {
		arr := utils.GenerateRandomSliceMinMax[byte](i, 0, 127)

		packed := Pack8to6(arr)
		unpacked := Unpack6to8(packed)

		if binStr(unpacked, false, 6) != binStr(packed, false, 8) {
			utils.Log(
				"arr", arr,
				"arr", binStr(arr, true, 8), "",
				"arr", binStr(arr, true, 6), "",
				"pac", binStr(packed, false, 8), "",
				"unp", binStr(unpacked, false, 6),
			)
			t.Fatal()
		}
	}
}

// binStr formats byte slices into binary strings with different modes
func binStr(data []byte, spacemode bool, bitmode int) string {
	var sb strings.Builder

	for _, b := range data {
		var format string
		var masked byte

		switch bitmode {
		case 4:
			format = "%04b"
			masked = b & 0x0F // Keep only the lower 4 bits
		case 5:
			format = "%05b"
			masked = b & 0x1F // Keep only the lower 5 bits
		case 6:
			format = "%06b"
			masked = b & 0x3F // Keep only the lower 6 bits
		default:
			format = "%08b"
			masked = b // Full byte
		}

		if spacemode {
			sb.WriteString(fmt.Sprintf(format+" ", masked))
		} else {
			sb.WriteString(fmt.Sprintf(format, masked))
		}
	}

	return strings.TrimSpace(sb.String()) // Trim trailing space if added
}
