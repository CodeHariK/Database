package encode

import (
	"strings"

	"github.com/codeharik/secretary/utils"
)

type SECKEY struct {
	sec  byte
	keys []byte
}

var SEC32KeyMap = [32]SECKEY{
	/*00*/ {'-', []byte{'~'}},
	/*01*/ {'a', []byte{'a', 'A', '@'}},
	/*02*/ {'b', []byte{'b', 'B', 'p', 'P', '+'}},
	/*03*/ {'c', []byte{'c', 'C', 'k', 'K', 'q', 'Q'}},
	/*04*/ {'d', []byte{'d', 'D', '/'}},
	/*05*/ {'e', []byte{'e', 'E', '='}},
	/*06*/ {'f', []byte{'f', 'F', '$'}},
	/*07*/ {'g', []byte{'g', 'G', 'j', 'J', 'z', 'Z'}},
	/*08*/ {'h', []byte{'h', 'H', '#'}},
	/*09*/ {'i', []byte{'i', 'I', '?', '!'}},
	/*10*/ {'l', []byte{'l', 'L', '^'}},
	/*11*/ {'m', []byte{'m', 'M', '*'}},
	/*12*/ {'n', []byte{'n', 'N', '-'}},
	/*13*/ {'o', []byte{'o', 'O', 'u', 'U'}},
	/*14*/ {'r', []byte{'r', 'R'}},
	/*15*/ {'s', []byte{'s', 'S', 'x', 'X'}},
	/*16*/ {'t', []byte{'t', 'T'}},
	/*17*/ {'v', []byte{'v', 'V', 'w', 'W'}},
	/*18*/ {'y', []byte{'y', 'Y', '&', '%'}},
	/*19*/ {'0', []byte{'0'}},
	/*20*/ {'1', []byte{'1'}},
	/*21*/ {'2', []byte{'2'}},
	/*22*/ {'3', []byte{'3'}},
	/*23*/ {'4', []byte{'4'}},
	/*24*/ {'5', []byte{'5'}},
	/*25*/ {'6', []byte{'6'}},
	/*26*/ {'7', []byte{'7'}},
	/*27*/ {'8', []byte{'8'}},
	/*28*/ {'9', []byte{'9'}},
	/*29*/ {'Q', []byte{'(', '[', '<', '{'}},
	/*30*/ {'R', []byte{')', ']', '>', '}'}},
	/*31*/ {'_', []byte{'_', '.', ',', ';', ' ', '\n', '"', '\'', '`', ':'}},
}

var (
	ASCII32 = [32]byte{126, 97, 98, 99, 100, 101, 102, 103, 104, 105, 108, 109, 110, 111, 114, 115, 116, 118, 121, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 40, 41, 95}
	SEC32   = [32]byte{45, 97, 98, 99, 100, 101, 102, 103, 104, 105, 108, 109, 110, 111, 114, 115, 116, 118, 121, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 81, 82, 95}

	ASCII32Index = [256]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 31, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 31, 9, 31, 8, 6, 18, 18, 31, 29, 30, 11, 2, 31, 12, 31, 4, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 31, 31, 29, 5, 30, 9, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 7, 3, 10, 11, 12, 13, 2, 3, 14, 15, 16, 13, 17, 17, 15, 18, 7, 29, 0, 30, 10, 31, 31, 1, 2, 3, 4, 5, 6, 7, 8, 9, 7, 3, 10, 11, 12, 13, 2, 3, 14, 15, 16, 13, 17, 17, 15, 18, 7, 29, 0, 30, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	SEC32Index   = [256]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 29, 30, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 31, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 0, 10, 11, 12, 13, 0, 0, 14, 15, 16, 0, 17, 0, 0, 18, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

// func init() {
// 	for i := range ASCII32Index {
// 		ASCII32Index[i] = 0
// 		SEC32Index[i] = 0
// 	}
// 	for index, chars := range SEC32KeyMap {
// 		ASCII32[index] = chars.asc[0]
// 		SEC32[index] = chars.sec
// 		SEC32Index[SEC32[index]] = byte(index)
// 		for _, c := range chars.asc {
// 			ASCII32Index[c] = byte(index)
// 		}
// 	}
// }

func StringToSec32(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC32[ASCII32Index[c]]
		}))
}

func StringToIndex32(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return ASCII32Index[c]
		})
}

func StringToIndex32Packed(str string) []byte {
	return Pack8to5(StringToIndex32(str))
}

func Index32ToAscii32(indexes []byte) string {
	return string(utils.Map(
		indexes,
		func(i byte) byte {
			return ASCII32[i]
		}))
}

func Index32PackedToAscii32(indexes []byte) string {
	return Index32ToAscii32(Unpack5to8(indexes))
}

func Sec32ToAscii32(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return ASCII32[SEC32Index[c]]
		}))
}

func Sec32ToIndex32(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC32Index[c]
		})
}

func Sec32ToIndex32Packed(str string) []byte {
	return Pack8to5(Sec32ToIndex32(str))
}

func Index32ToSec32(indexes []byte) string {
	return string(utils.Map(
		indexes,
		func(i byte) byte {
			return SEC32[i]
		}))
}

func Index32PackedToSec32(indexes []byte) string {
	return Index32ToSec32(Unpack5to8(indexes))
}

func ExpandStringToSec32(str string) string {
	unpacked := Unpack5to8([]byte(str))
	enc := make([]byte, len(unpacked))
	for i := 0; i < len(unpacked); i++ {
		enc[i] = SEC32[unpacked[i]]
	}
	return string(enc)
}

func Sec32ToExpandString(str string) string {
	packed := Pack8to5(Sec32ToIndex32(str))
	return strings.Trim(string(packed), "\x00")
}

// 87654321 | 87654321 | 87654321 | 87654321 | 87654321 | 87654321 | 87654321 | 87654321
// Encode
// 0,5   1,3  1,2 2,5 3,1   3,4  4,4   4,1 5,5 6,2   6,3  7,5
// 54321 543 | 21 54321 5 | 4321 5432 | 1 54321 54 | 321 54321
// Pack8to5 converts 8-bit bytes into 5-bit packed format
func Pack8to5(input []byte) []byte {
	// Calculate the number of bytes needed for the packed output
	// Each 8 bits (1 byte) will produce 5 bits, so we need to round up
	// the number of input bytes to the nearest multiple of 8.
	e := len(input) % 8
	if e != 0 {
		input = append(input, make([]byte, 8-e)...)
	}

	// The length of the packed output will be (5/8) * len(input)
	packed := make([]byte, (5*len(input))/8)

	for u, p := 0, 0; u < len(input); u += 8 {
		packed[p] = (input[u]&0b00011111)<<3 | (input[u+1]&0b00011100)>>2
		packed[p+1] = (input[u+1]&0b00000011)<<6 | (input[u+2]&0b00011111)<<1 | (input[u+3]&0b00010000)>>4
		packed[p+2] = input[u+3]<<4 | (input[u+4]&0b00011110)>>1
		packed[p+3] = (input[u+4]&0b00000001)<<7 | (input[u+5]&0b00011111)<<2 | (input[u+6]&0b00011000)>>3
		packed[p+4] = input[u+6]<<5 | (input[u+7] & 0b00011111)
		p += 5
	}

	return packed
}

// Encode
// 54321 543 | 21 54321 5 | 4321 5432 | 1 54321 54 | 321 54321
// Decode
// 0,5        0,3  1,2   1,5        1,1  2,4   2,4  3,1   3,5        3,2  4,3   4,5
// 00054321 | 00054321 | 00054321 | 00054321 | 00054321 | 00054321 | 00054321 | 00054321
func Unpack5to8(packed []byte) []byte {
	e := len(packed) % 5
	if e != 0 {
		packed = append(packed, make([]byte, 5-e)...)
	}

	unpacked := make([]byte, (8*len(packed))/5)

	for u, p := 0, 0; p < len(packed); p += 5 {
		unpacked[u] = packed[p] >> 3
		unpacked[u+1] = (packed[p]<<2 | packed[p+1]>>6) & 0b00011111
		unpacked[u+2] = (packed[p+1] >> 1) & 0b00011111
		unpacked[u+3] = (packed[p+1]<<4 | packed[p+2]>>4) & 0b00011111
		unpacked[u+4] = (packed[p+2]<<1 | packed[p+3]>>7) & 0b00011111
		unpacked[u+5] = (packed[p+3] >> 2) & 0b00011111
		unpacked[u+6] = (packed[p+3]<<3 | packed[p+4]>>5) & 0b00011111
		unpacked[u+7] = packed[p+4] & 0b00011111
		u += 8
	}

	return unpacked
}
