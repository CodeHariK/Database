package encode

import (
	"strings"

	"github.com/codeharik/secretary/utils"
)

var SEC32KeyMap = [][]byte{
	/*00*/ {'~'},
	/*01*/ {'a', 'A', '@'},
	/*02*/ {'b', 'B', 'p', 'P', '+'},
	/*03*/ {'c', 'C', 'k', 'K', 'q', 'Q'},
	/*04*/ {'d', 'D', '/'},
	/*05*/ {'e', 'E', '='},
	/*06*/ {'f', 'F', '$'},
	/*07*/ {'g', 'G', 'j', 'J', 'z', 'Z'},
	/*08*/ {'h', 'H', '#'},
	/*09*/ {'i', 'I', '?', '!'},
	/*10*/ {'l', 'L', '^'},
	/*11*/ {'m', 'M', '*'},
	/*12*/ {'n', 'N', '-'},
	/*13*/ {'o', 'O', 'u', 'U'},
	/*14*/ {'r', 'R'},
	/*15*/ {'s', 'S', 'x', 'X'},
	/*16*/ {'t', 'T'},
	/*17*/ {'v', 'V', 'w', 'W'},
	/*18*/ {'y', 'Y', '&', '%'},
	/*19*/ {'0'},
	/*20*/ {'1'},
	/*21*/ {'2'},
	/*22*/ {'3'},
	/*23*/ {'4'},
	/*24*/ {'5'},
	/*25*/ {'6'},
	/*26*/ {'7'},
	/*27*/ {'8'},
	/*28*/ {'9'},
	/*29*/ {'(', '[', '<', '{'},
	/*30*/ {')', ']', '>', '}'},
	/*31*/ {'_', '.', ',', ';', ' ', '\n', '"', '\'', '`', ':'},
}

var (
	ASCII32 = [32]byte{}
	SEC32   = []byte(`-abcdefghilmnorstvy0123456789QR_`)
)

var (
	ASCII32Index = [256]byte{}
	SEC32Index   = [256]byte{}
)

func init() {
	for i := range ASCII32Index {
		ASCII32Index[i] = 0
		SEC32Index[i] = 0
	}
	for index, chars := range SEC32KeyMap {
		ASCII32[index] = chars[0]
		SEC32Index[SEC32[index]] = byte(index)
		for _, c := range chars {
			ASCII32Index[c] = byte(index)
		}
	}
}

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
