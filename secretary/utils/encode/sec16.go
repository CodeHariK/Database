package encode

import (
	"strings"

	"github.com/codeharik/secretary/utils"
)

var SEC16KeyMap = [16]SECKEY{
	/*00*/ {'-', []byte{'-'}},
	/*01*/ {'a', []byte{'a', 'A', 'h', 'H'}},
	/*02*/ {'p', []byte{'p', 'P', 'b', 'B'}},
	/*03*/ {'c', []byte{'c', 'C', 'k', 'K', 'q', 'Q'}},
	/*04*/ {'t', []byte{'t', 'T', 'd', 'D'}},
	/*05*/ {'e', []byte{'e', 'E', 'i', 'I'}},
	/*06*/ {'f', []byte{'f', 'F'}},
	/*07*/ {'g', []byte{'g', 'G', 'j', 'J', 'z', 'Z'}},
	/*08*/ {'l', []byte{'l', 'L'}},
	/*09*/ {'n', []byte{'n', 'N', 'm', 'M'}},
	/*10*/ {'o', []byte{'o', 'O', 'u', 'U'}},
	/*11*/ {'r', []byte{'r', 'R'}},
	/*12*/ {'s', []byte{'s', 'S', 'x', 'X'}},
	/*13*/ {'y', []byte{'y', 'Y', 'v', 'V', 'w', 'W'}},
	/*14*/ {'0', []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}},
	/*15*/ {'_', []byte{'_', '.', ',', ';', ' ', '\n', '"', '\'', '`', ':'}},
}

var (
	ASCII16 = [16]byte{45, 97, 112, 99, 116, 101, 102, 103, 108, 110, 111, 114, 115, 121, 48, 95}
	SEC16   = [16]byte{45, 97, 112, 99, 116, 101, 102, 103, 108, 110, 111, 114, 115, 121, 48, 95}

	ASCII16Index = [256]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 15, 0, 15, 0, 0, 0, 0, 15, 0, 0, 0, 0, 15, 0, 15, 0, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 15, 15, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 1, 5, 7, 3, 8, 9, 9, 10, 2, 3, 11, 12, 4, 10, 13, 13, 12, 13, 7, 0, 0, 0, 0, 15, 15, 1, 2, 3, 4, 5, 6, 7, 1, 5, 7, 3, 8, 9, 9, 10, 2, 3, 11, 12, 4, 10, 13, 13, 12, 13, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	SEC16Index   = [256]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 14, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 15, 0, 1, 0, 3, 0, 5, 6, 7, 0, 0, 0, 0, 8, 0, 9, 10, 2, 0, 11, 12, 4, 0, 0, 0, 0, 13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

// func init() {
// 	for i := range ASCII16Index {
// 		ASCII16Index[i] = 0
// 		SEC16Index[i] = 0
// 	}
// 	for index, chars := range SEC16KeyMap {
// 		ASCII16[index] = chars.keys[0]
// 		SEC16[index] = chars.sec
// 		SEC16Index[SEC16[index]] = byte(index)
// 		for _, c := range chars.keys {
// 			ASCII16Index[c] = byte(index)
// 		}
// 	}

// 	fmt.Println(utils.Map(SEC16Index[:], func(a byte) string { return fmt.Sprint(a, ",") }))
// }

func StringToSec16(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC16[ASCII16Index[c]]
		}))
}

func StringToIndex16(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return ASCII16Index[c]
		})
}

func StringToIndex16Packed(str string) []byte {
	return Pack8to4(StringToIndex16(str))
}

func Index16ToAscii16(indexes []byte) string {
	return string(utils.Map(
		indexes,
		func(i byte) byte {
			return ASCII16[i]
		}))
}

func Index16PackedToAscii16(indexes []byte) string {
	return Index16ToAscii16(Unpack4to8(indexes))
}

func Sec16ToAscii16(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return ASCII16[SEC16Index[c]]
		}))
}

func Sec16ToIndex16(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC16Index[c]
		})
}

func Sec16ToIndex16Packed(str string) []byte {
	return Pack8to4(Sec16ToIndex16(str))
}

func Index16ToSec16(indexes []byte) string {
	return string(utils.Map(
		indexes,
		func(i byte) byte {
			return SEC16[i]
		}))
}

func Index16PackedToSec16(indexes []byte) string {
	return Index16ToSec16(Unpack4to8(indexes))
}

func ExpandStringToSec16(str string) string {
	unpacked := Unpack4to8([]byte(str))
	enc := make([]byte, len(unpacked))
	for i := 0; i < len(unpacked); i++ {
		enc[i] = SEC16[unpacked[i]]
	}
	return string(enc)
}

func Sec16ToExpandString(str string) string {
	packed := Pack8to4(Sec16ToIndex16(str))
	return strings.Trim(string(packed), "\x00")
}

func Pack8to4(input []byte) []byte {
	e := len(input) % 2
	if e != 0 {
		input = append(input, 0)
	}

	packed := make([]byte, len(input)/2)
	for u, p := 0, 0; u < len(input); u += 2 {
		packed[p] = (input[u] << 4) | (input[u+1] & 0x0F)
		p++
	}
	return packed
}

func Unpack4to8(packed []byte) []byte {
	unpacked := make([]byte, (2 * len(packed)))

	for u, p := 0, 0; p < len(packed); p += 1 {
		unpacked[u] = packed[p] >> 4
		unpacked[u+1] = packed[p] & 0b00001111
		u += 2
	}

	return unpacked
}
