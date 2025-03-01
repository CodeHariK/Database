package encode

import (
	"io"
	"strings"

	"github.com/codeharik/secretary/utils"
)

var SEC64KeyMap = [64]SECKEY{
	/*00*/ {'-', []byte{'~'}},
	/*01*/ {'a', []byte{'a', 'A'}},
	/*02*/ {'b', []byte{'b', 'B'}},
	/*03*/ {'c', []byte{'c', 'C'}},
	/*04*/ {'d', []byte{'d', 'D'}},
	/*05*/ {'e', []byte{'e', 'E'}},
	/*06*/ {'f', []byte{'f', 'F'}},
	/*07*/ {'g', []byte{'g', 'G'}},
	/*08*/ {'h', []byte{'h', 'H'}},
	/*09*/ {'i', []byte{'i', 'I'}},
	/*10*/ {'j', []byte{'j', 'J'}},
	/*11*/ {'k', []byte{'k', 'K'}},
	/*12*/ {'l', []byte{'l', 'L'}},
	/*13*/ {'m', []byte{'m', 'M'}},
	/*14*/ {'n', []byte{'n', 'N'}},
	/*15*/ {'o', []byte{'o', 'O'}},
	/*16*/ {'p', []byte{'p', 'P'}},
	/*17*/ {'q', []byte{'q', 'Q'}},
	/*18*/ {'r', []byte{'r', 'R'}},
	/*19*/ {'s', []byte{'s', 'S'}},
	/*20*/ {'t', []byte{'t', 'T'}},
	/*21*/ {'u', []byte{'u', 'U'}},
	/*22*/ {'v', []byte{'v', 'V'}},
	/*23*/ {'w', []byte{'w', 'W'}},
	/*24*/ {'x', []byte{'x', 'X'}},
	/*25*/ {'y', []byte{'y', 'Y'}},
	/*26*/ {'z', []byte{'z', 'Z'}},
	/*27*/ {'0', []byte{'0'}},
	/*28*/ {'1', []byte{'1'}},
	/*29*/ {'2', []byte{'2'}},
	/*30*/ {'3', []byte{'3'}},
	/*31*/ {'4', []byte{'4'}},
	/*32*/ {'5', []byte{'5'}},
	/*33*/ {'6', []byte{'6'}},
	/*34*/ {'7', []byte{'7'}},
	/*35*/ {'8', []byte{'8'}},
	/*36*/ {'9', []byte{'9'}},
	/*37*/ {'A', []byte{'='}},
	/*38*/ {'B', []byte{'+'}},
	/*39*/ {'C', []byte{'-'}},
	/*40*/ {'D', []byte{'*'}},
	/*41*/ {'E', []byte{'/'}},
	/*42*/ {'F', []byte{'\\', '|'}},
	/*43*/ {'G', []byte{'%'}},
	/*44*/ {'H', []byte{'^'}},
	/*45*/ {'I', []byte{'<'}},
	/*46*/ {'J', []byte{'>'}},
	/*47*/ {'K', []byte{'!'}},
	/*48*/ {'L', []byte{'?'}},
	/*49*/ {'M', []byte{'@'}},
	/*50*/ {'N', []byte{'#'}},
	/*51*/ {'O', []byte{'$'}},
	/*52*/ {'P', []byte{'&'}},
	/*53*/ {'Q', []byte{'(', '[', '{'}},
	/*54*/ {'R', []byte{')', ']', '}'}},
	/*55*/ {'S', []byte{','}},
	/*56*/ {'T', []byte{';'}},
	/*57*/ {'U', []byte{':'}},
	/*58*/ {'V', []byte{'\''}},
	/*59*/ {'W', []byte{'"', '`'}},
	/*60*/ {'X', []byte{'_'}},
	/*61*/ {'.', []byte{'.'}},
	/*62*/ {'_', []byte{' '}},
	/*63*/ {'+', []byte{'\n'}},
}

var (
	ASCII64 = [64]byte{126, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 61, 43, 45, 42, 47, 92, 37, 94, 60, 62, 33, 63, 64, 35, 36, 38, 40, 41, 44, 59, 58, 39, 34, 95, 46, 32, 10}
	SEC64   = [64]byte{45, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 46, 95, 43}

	ASCII64Index = [256]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 63, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 62, 47, 59, 50, 51, 43, 52, 58, 53, 54, 40, 38, 55, 39, 61, 41, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 57, 56, 45, 37, 46, 48, 49, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 53, 42, 54, 44, 60, 59, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 53, 42, 54, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	SEC64Index   = [256]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 63, 0, 0, 61, 0, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 0, 0, 0, 0, 0, 0, 0, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 0, 0, 0, 0, 0, 0, 62, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

// func init() {
// 	for i := range ASCII64Index {
// 		ASCII64Index[i] = 0
// 		SEC64Index[i] = 0
// 	}

// 	for index, chars := range SEC64KeyMap {
// 		ASCII64[index] = chars.asc[0]
// 		SEC64[index] = chars.sec
// 		SEC64Index[SEC64[index]] = byte(index)
// 		for _, c := range chars.asc {
// 			ASCII64Index[c] = byte(index)
// 		}
// 	}
// }

func StringToSec64(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC64[ASCII64Index[c]]
		}))
}

func StringToIndex64(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return ASCII64Index[c]
		})
}

func StringToIndex64Packed(str string) []byte {
	return Pack8to6(StringToIndex64(str))
}

func Index64ToAscii64(indexes []byte) string {
	return string(utils.Map(
		indexes,
		func(i byte) byte {
			return ASCII64[i]
		}))
}

func Index64PackedToAscii64(indexes []byte) string {
	return Index64ToAscii64(Unpack6to8(indexes))
}

func Sec64ToAscii64(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return ASCII64[SEC64Index[c]]
		}))
}

func Sec64ToIndex64(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC64Index[c]
		})
}

func Sec64ToIndex64Packed(str string) []byte {
	return Pack8to6(Sec64ToIndex64(str))
}

func Index64ToSec64(indexes []byte) string {
	return string(utils.Map(
		indexes,
		func(i byte) byte {
			return SEC64[i]
		}))
}

func Index64PackedToSec64(indexes []byte) string {
	return Index64ToSec64(Unpack6to8(indexes))
}

func ExpandStringToSec64(str string) string {
	unpacked := Unpack6to8([]byte(str))
	enc := make([]byte, len(unpacked))
	for i := 0; i < len(unpacked); i++ {
		enc[i] = SEC64[unpacked[i]]
	}
	return string(enc)
}

func Sec64ToExpandString(str string) string {
	packed := Pack8to6(Sec64ToIndex64(str))
	return strings.Trim(string(packed), "\x00")
}

// 87654321 | 87654321 | 87654321 | 87654321
// Encode
// 0,6    1,2  1,4  3,4   3,2 4,6
// 654321 65 | 4321 6543 | 21 654321
// Pack8to6 converts 4 bytes (8-bit each) into 3 bytes (6-bit each)
func Pack8to6(input []byte) []byte {
	e := len(input) % 4
	if e != 0 {
		input = append(input, make([]byte, 4-e)...)
	}

	packed := make([]byte, (3*len(input))/4)

	for u, p := 0, 0; u < len(input); u += 4 {
		// 6 bits of byte 1, 2 bits from byte 2
		packed[p] = (input[u] << 2) | ((input[u+1] & 0b00110000) >> 4)
		// 4 bits of byte 2, 4 bits from byte 3
		packed[p+1] = ((input[u+1] & 0b00001111) << 4) | ((input[u+2] & 0b00111100) >> 2)
		// 2 bits of byte 3, 6 bits from byte 4
		packed[p+2] = (input[u+2] << 6) | (input[u+3] & 0b00111111)
		p += 3
	}

	return packed
}

// Encode
// 654321 65 | 4321 6543 | 21 654321
// Decode
// 0,6        0,2  1,4   1,4  2,2   2,6
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

// Sec64BufferedEncoder is a buffered encoder similar to base64.NewEncoder.
type Sec64BufferedEncoder struct {
	w      io.Writer // Underlying writer
	buffer []byte    // Internal buffer
}

// NewCustomBufferedEncoder creates a new buffered encoder.
func NewSec64BufferedEncoder(w io.Writer) io.WriteCloser {
	return &Sec64BufferedEncoder{
		w:      w,
		buffer: make([]byte, 0, 64), // Example buffer size of 64 bytes
	}
}

// Write encodes data and writes it in chunks.
func (e *Sec64BufferedEncoder) Write(p []byte) (n int, err error) {
	e.buffer = append(e.buffer, p...) // Buffer the data

	// Simulating encoding and writing in chunks
	for len(e.buffer) >= 3 { // Example: Encoding works in chunks of 3
		encoded := ExpandStringToSec64(string(e.buffer[:3])) // Encode a chunk
		_, err := e.w.Write([]byte(encoded))                 // Write to underlying writer
		if err != nil {
			return 0, err
		}
		e.buffer = e.buffer[3:] // Remove written chunk
	}
	return len(p), nil
}

// Close flushes any remaining buffered data.
func (e *Sec64BufferedEncoder) Close() error {
	if len(e.buffer) > 0 {
		encoded := ExpandStringToSec64(string(e.buffer)) // Encode remaining data
		_, err := e.w.Write([]byte(encoded))
		if err != nil {
			return err
		}
		e.buffer = nil // Clear buffer
	}
	return nil
}
