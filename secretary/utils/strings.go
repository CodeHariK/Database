package utils

import (
	"fmt"
	"math/rand"
	"regexp"
)

var SAFE_COLLECTION_REGEX = regexp.MustCompile(`[^a-zA-Z0-9._]`)

// SafeCollectionString removes all characters except letters and digits, lowercase
func SafeCollectionString(input string) string {
	return SAFE_COLLECTION_REGEX.ReplaceAllString(input, "")
}

// GenerateRandomString generates a random string of given length `l`
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func GenerateSeqString(sequence *uint64, length int) string {
	*sequence += 1
	return fmt.Sprintf("%0*d", length, *sequence)
}

func GenerateSeqRandomString(sequence *uint64, length int, pad int, value ...string) string {
	*sequence += 1
	str := fmt.Sprintf("%0*d:%s:%s", pad, *sequence, value, GenerateRandomString(length))
	return str[:length]
}

// Lossy encoding

const (
	SEC64 = `abcdefghijklmnopqrstuvwxyz0123456789=+-*/\%^<>!?@#$&(),;:'"_.SNQ`
	ENC64 = `abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMOPRTUVWXYZ_.SNQ`
)

// Asc   Sec  	Enc
// a     a		a
// A     a		a
// !     !		K
// K     k		k
// {     (		T

/**
*      Character Type							Mapping
*      A-Z,a-z									a-z
*      0-9										0-9
*      Symbols (=+-*\/\%^<>!?@#$&(),;:'"_.)		ABCDEFGHIJKLMOPRTUVWXYZ_.
*      (space)									S
*      \n (newline)								N
*      Any other character						Q
 */

var (
	ASCIIToSEC64Map [256]byte
	ASCIIToENCMap   [256]byte
	SEC64ToENCMap   [256]byte
	ENCToSEC64Map   [256]byte
)

func init() {
	for i := range ENCToSEC64Map {
		SEC64ToENCMap[i] = 0
		ENCToSEC64Map[i] = 'Q'
		ASCIIToSEC64Map[i] = 'Q'
		ASCIIToENCMap[i] = 'Q'
	}

	// Initialize encoding map
	for i, c := range SEC64 {
		SEC64ToENCMap[c] = ENC64[i]
		ENCToSEC64Map[ENC64[i]] = byte(c)
		ASCIIToSEC64Map[c] = byte(c)
		ASCIIToENCMap[c] = ENC64[i]
	}

	// Map A-Z to a-z
	for c := 'A'; c <= 'Z'; c++ {
		ASCIIToSEC64Map[c] = SEC64[c-'A']
		ASCIIToENCMap[c] = SEC64[c-'A']
	}

	// Map brackets to ()
	brackets := map[rune]rune{'[': '(', ']': ')', '{': '(', '}': ')'}
	for k, v := range brackets {
		ASCIIToSEC64Map[k] = ASCIIToSEC64Map[v]
		ASCIIToENCMap[k] = ASCIIToENCMap[v]
	}

	ASCIIToSEC64Map[' '] = 'S'  // Space to 'S'
	ASCIIToSEC64Map['\n'] = 'N' // Newline to 'N'
	ASCIIToSEC64Map[0] = 'Q'    // Null to 'Q'

	ASCIIToENCMap[' '] = 'S'  // Space to 'S'
	ASCIIToENCMap['\n'] = 'N' // Newline to 'N'
	ASCIIToENCMap[0] = 'Q'    // Null to 'Q'

	for i := 0; i < 128; i++ {
		c := rune(i)
		fmt.Printf(
			"%3d  %-7q  S2E: %-7q  E2S: %-7q  A2S: %-7q A2E: %-7q\n",
			i, c,
			SEC64ToENCMap[c], ENCToSEC64Map[c], ASCIIToSEC64Map[c], ASCIIToENCMap[c],
		)
	}
}

func AsciiToEncEncode(ascii string) string {
	enc64 := ""
	for i := 0; i < len(ascii); i++ {
		enc64 += string(ASCIIToENCMap[ascii[i]])
	}
	return enc64
}

func EncToSEC64Decode(enc64 string) string {
	sec64 := ""
	for i := 0; i < len(enc64); i++ {
		sec64 += string(ENCToSEC64Map[enc64[i]])
	}
	return sec64
}

// func SEC64ToAsciiDecode(sec64 string) string {
// 	ascii := ""
// 	for i := 0; i < len(ascii); i++ {
// 		ascii += string(Sec64[ascii[i]])
// 	}
// 	return ascii
// }

// func encode(input []byte) string {
// 	outputLen := (len(input) * 8) / 6
// 	var output strings.Builder
// 	output.Grow(outputLen)
// 	var buffer uint32
// 	var bits uint

// 	for _, b := range input {
// 		buffer = (buffer << 8) | uint32(b)
// 		bits += 8
// 		for bits >= 6 {
// 			bits -= 6
// 			output.WriteByte(SEC64[(buffer>>bits)&0x3F])
// 		}
// 	}

// 	// Handle remaining bits (if any)
// 	if bits > 0 {
// 		output.WriteByte(SEC64[(buffer<<(6-bits))&0x3F])
// 	}

// 	return output.String()
// }

// func decode(input string) ([]byte, error) {
// 	outputLen := (len(input) * 6) / 8
// 	output := make([]byte, outputLen)
// 	var buffer uint32
// 	var bits uint
// 	var index int

// 	for _, c := range input {
// 		value := SEC64ToASCIIMap[c]
// 		if value == 0 && c == 'E' {
// 			continue // Skip null characters
// 		}
// 		buffer = (buffer << 6) | uint32(value)
// 		bits += 6
// 		for bits >= 8 {
// 			bits -= 8
// 			output[index] = byte((buffer >> bits) & 0xFF)
// 			index++
// 		}
// 	}

// 	return output, nil
// }
