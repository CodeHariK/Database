package encode

import (
	"io"
	"strings"

	"github.com/codeharik/secretary/utils"
)

var (
	/**
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

	//		           ABCDEFGHIJKLMNOPQRSTUVWXYZ               |         [{}]   `
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
	brackets := map[rune]rune{'[': '(', ']': ')', '{': '(', '}': ')', '|': '\\', '`': '"'}
	for k, v := range brackets {
		ASCII64Index[k] = ASCII64Index[v]
	}
}

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
