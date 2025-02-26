package sec64

import (
	"fmt"
	"strings"
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
	ASCII = []byte(`~abcdefghijklmnopqrstuvwxyz0123456789=+-*/\%^<>!?@#$&(),;:'"_. N`)
	SEC64 = []byte(`-abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWX._+`)
	//              0123456789012345678901234567890123456789012345678901234567890123
)

type Sec64 struct {
	index byte
	char  byte
}

var (
	Ascii2Sec [256]Sec64
	Sec2Ascii [256]Sec64
)

func init() {
	ASCII[63] = '\n'

	for i := range Ascii2Sec {
		Sec2Ascii[i] = Sec64{index: 0, char: '~'}
		Ascii2Sec[i] = Sec64{index: 0, char: '-'}
	}
	for i := 1; i < 63; i++ {
		Ascii2Sec[ASCII[i]] = Sec64{index: byte(i), char: SEC64[i]}
		Sec2Ascii[SEC64[i]] = Sec64{index: byte(i), char: ASCII[i]}
	}
	for c := 'A'; c <= 'Z'; c++ {
		Ascii2Sec[c] = Sec64{index: byte(c - 'A' + 1), char: SEC64[c-'A'+1]}
	}
	brackets := map[rune]rune{'[': '(', ']': ')', '{': '(', '}': ')', '|': '\\'}
	for k, v := range brackets {
		Ascii2Sec[k] = Ascii2Sec[v]
	}

	Sec2Ascii['+'] = Sec64{index: 63, char: '\n'}
	Ascii2Sec['\n'] = Sec64{index: 63, char: '+'}

	for i := 0; i < 128; i++ {
		c := rune(i)
		fmt.Printf(
			"%3d %-7q  A2S: %-4d %-4q     S2A: %-4d  %-4q\n",
			i, c,
			Ascii2Sec[c].index, Ascii2Sec[c].char, Sec2Ascii[c].index, Sec2Ascii[c].char,
		)
	}
	fmt.Println()
	fmt.Println(string(SEC64))
	fmt.Println(string(ASCII))
}

func AsciiToSec64(str string) string {
	enc := make([]byte, len(str))
	for i := 0; i < len(str); i++ {
		enc[i] = Ascii2Sec[str[i]].char
	}
	return string(enc)
}

func AsciiToIndex(str string) []byte {
	enc := make([]byte, len(str))
	for i := 0; i < len(str); i++ {
		enc[i] = Ascii2Sec[str[i]].index
	}
	return enc
}

func IndexToAscii(indexes []byte) string {
	str := make([]byte, len(indexes))
	for i := 0; i < len(indexes); i++ {
		str[i] = ASCII[indexes[i]]
		// fmt.Printf("%-3d %-3d %-3q %-3d\n", i, indexes[i], string(str[i]), str[i])
	}
	fmt.Println()
	return string(str)
}

func Sec64ToAscii(str string) string {
	dec := make([]byte, len(str))
	for i := 0; i < len(str); i++ {
		dec[i] = Sec2Ascii[str[i]].char
	}
	return string(dec)
}

func Sec64ToIndex(str string) []byte {
	dec := make([]byte, len(str))
	for i := 0; i < len(str); i++ {
		dec[i] = Sec2Ascii[str[i]].index
	}
	return dec
}

func IndexToSec64(indexes []byte) string {
	str := make([]byte, len(indexes))
	for i := 0; i < len(indexes); i++ {
		str[i] = SEC64[indexes[i]]
	}
	return string(str)
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
