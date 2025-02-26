package encode

import (
	"strings"

	"github.com/codeharik/secretary/utils"
)

var (
	/**
	*
	*		Url safe Lossy encoding : 8bit -> 6bit
	*
	*		Character Type							Mapping
	*		A-Z,a-z									a-z
	*		0-9										0-9
	*		Symbols (=+-*\/\%^<>!?@#$&(),;:'"_.)	ABCDEFGHIJKLMOPQRTUVWXYZ.
	*		(space)									S
	*		\n (newline)							N
	*		Any other character						_
	**/

	//		         ABCDEFGHIJKLMNOPQRSTUVWXYZ               |         [{}]
	ASCII64 = []byte(`~abcdefghijklmnopqrstuvwxyz0123456789=+-*/\%^<>!?@#$&(),;:'"_. N`)
	SEC64   = []byte(`-abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWX._+`)
	//              0123456789012345678901234567890123456789012345678901234567890123
)

var (
	ASCII64Index = [256]byte{}
	SEC64Index   = [256]byte{}
)

func init() {
	ASCII64[63] = '\n'

	for i := range ASCII64Index {
		ASCII64Index[i] = 0
		SEC64Index[i] = 0
	}

	for i := 1; i < 64; i++ {
		ASCII64Index[ASCII64[i]] = byte(i)
		SEC64Index[SEC64[i]] = byte(i)
	}
	for c := 'A'; c <= 'Z'; c++ {
		ASCII64Index[c] = byte(c - 'A' + 1)
	}
	brackets := map[rune]rune{'[': '(', ']': ')', '{': '(', '}': ')', '|': '\\'}
	for k, v := range brackets {
		ASCII64Index[k] = ASCII64Index[v]
	}
}

func AsciiToSec64(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC64[ASCII64Index[c]]
		}))
}

func Ascii64ToIndex(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return ASCII64Index[c]
		})
}

func IndexToAscii64(indexes []byte) string {
	return string(utils.Map(
		indexes,
		func(i byte) byte {
			return ASCII64[i]
		}))
}

func Sec64ToAscii(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return ASCII64[SEC64Index[c]]
		}))
}

func Sec64ToIndex(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC64Index[c]
		})
}

func IndexToSec64(indexes []byte) string {
	return string(utils.Map(
		indexes,
		func(i byte) byte {
			return SEC64[i]
		}))
}

func AsciiToSec64Expand(str string) string {
	unpacked := Unpack6to8([]byte(str))
	enc := make([]byte, len(unpacked))
	for i := 0; i < len(unpacked); i++ {
		enc[i] = SEC64[unpacked[i]]
	}
	return string(enc)
}

func Sec64ToAsciiExpand(str string) string {
	packed := Pack8to6(Sec64ToIndex(str))
	return strings.Trim(string(packed), "\x00")
}

// 87654321 | 87654321 | 87654321 | 87654321
// Encode
// 65432165 | 43216543 | 21654321
// Pack8to6 converts 4 bytes (8-bit each) into 3 bytes (6-bit each)
func Pack8to6(input []byte) []byte {
	e := len(input) % 4
	if e != 0 {
		input = append(input, make([]byte, 4-e)...)
	}

	packed := make([]byte, (3*len(input))/4)

	for u, p := 0, 0; u < len(input); u += 4 {
		// 6 bits of byte 1, 2 bits from byte 2
		packed[p] = (input[u] << 2) | ((input[u+1] >> 4) & 0b00000011)
		// 4 bits of byte 2, 4 bits from byte 3
		packed[p+1] = ((input[u+1] & 0b00001111) << 4) | ((input[u+2] >> 2) & 0b00001111)
		// 2 bits of byte 3, 6 bits from byte 4
		packed[p+2] = (input[u+2] << 6) | (input[u+3] & 0b00111111)
		p += 3
	}

	return packed
}

// Encode
// 65432165 | 43216543 | 21654321
// Decode
// 00654321 | 00654321 | 00654321 | 00654321
// Unpack6to8 converts 6-bit packed slices back to 8-bit byte slices (4-value → 3-byte chunks)
func Unpack6to8(packed []byte) []byte {
	e := len(packed) % 3
	if e != 0 {
		packed = append(packed, make([]byte, 3-e)...)
	}

	unpacked := make([]byte, (4*len(packed))/3)

	for u, p := 0, 0; p < len(packed); p += 3 {
		unpacked[u] = packed[p] >> 2
		unpacked[u+1] = (packed[p]<<4 | packed[p+1]>>4) & 0b00111111
		unpacked[u+2] = (packed[p+1]<<2 | packed[p+2]>>6) & 0b00111111
		unpacked[u+3] = packed[p+2] & 0b00111111
		u += 4
	}

	return unpacked
}
