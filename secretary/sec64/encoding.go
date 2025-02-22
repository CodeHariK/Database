package sec64

const (
	/**

			Url safe Lossy encoding : 8bit -> 6bit

	*		Character Type							Mapping
	*		A-Z,a-z									a-z
	*		0-9										0-9
	*		Symbols (=+-*\/\%^<>!?@#$&(),;:'"_.)	ABCDEFGHIJKLMOPRTUVWXYZ_.
	*		(space)									S
	*		\n (newline)							N
	*		Any other character						_
	**/

	//		    ABCDEFGHIJKLMNOPQRSTUVWXYZ                         [{}]
	ASCII = `_SNabcdefghijklmnopqrstuvwxyz0123456789=+-*/\%^<>!?@#$&(),;:'"_.`
	SEC64 = `_SNabcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMOPQRTUVWXYZ.`
)

type Sec64 struct {
	index int8
	char  byte
}

var (
	Ascii2Sec [256]Sec64
	Sec2Ascii [256]Sec64
)

func init() {
	for i := range Ascii2Sec {
		Ascii2Sec[i] = Sec64{index: 0, char: '_'}
		Sec2Ascii[i] = Sec64{index: 0, char: '\x00'}
	}
	for i := 0; i < 64; i++ {
		Ascii2Sec[ASCII[i]] = Sec64{index: int8(i), char: SEC64[i]}
		Sec2Ascii[SEC64[i]] = Sec64{index: int8(i), char: ASCII[i]}
	}
	for c := 'A'; c <= 'Z'; c++ {
		Ascii2Sec[c] = Sec64{index: int8(c - 'A' + 3), char: SEC64[c-'A'+3]}
	}
	brackets := map[rune]rune{'[': '(', ']': ')', '{': '(', '}': ')'}
	for k, v := range brackets {
		Ascii2Sec[k] = Ascii2Sec[v]
	}

	Ascii2Sec[' '] = Sec64{index: 1, char: 'S'}  // Space to 'S'
	Ascii2Sec['\n'] = Sec64{index: 2, char: 'N'} // Newline to 'N'

	Sec2Ascii['S'] = Sec64{index: 1, char: ' '}    // Space to 'S'
	Sec2Ascii['N'] = Sec64{index: 2, char: '\n'}   // Newline to 'N'
	Sec2Ascii['_'] = Sec64{index: 0, char: '\x00'} // Null to '_'

	// for i := 0; i < 128; i++ {
	// 	c := rune(i)
	// 	fmt.Printf(
	// 		"%3d %-7q  A2S:%-7q  S2A:%-7q\n",
	// 		i, c,
	// 		Ascii2Sec[c], Sec2Ascii[c],
	// 	)
	// }
}

func EncodeString(ascii string) string {
	enc := make([]byte, len(ascii))
	for i := 0; i < len(ascii); i++ {
		enc[i] = Ascii2Sec[ascii[i]].char
	}
	return string(enc)
}

func DecodeString(enc string) string {
	ascii := make([]byte, len(enc))
	for i := 0; i < len(enc); i++ {
		ascii[i] += Sec2Ascii[enc[i]].char
	}
	return string(ascii)
}

func Encode(ascii string) []byte {
	hello := []byte(EncodeString(ascii))
	return Pack8to6(hello)
}

func Decode(enc []byte) string {
	hello := Unpack6to8(enc)
	return DecodeString(string(hello))
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
