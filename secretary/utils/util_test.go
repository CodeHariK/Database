package utils

import (
	"reflect"
	"testing"
)

func TestStringsToArray(t *testing.T) {
	var keySeq uint64 = 0
	var sortedKeys [][]byte

	for r := 0; r < 26; r++ {
		key := []byte(GenerateSeqString(&keySeq, 16))
		sortedKeys = append(sortedKeys, key)
	}

	shuffledKeys := make([][][]byte, 4)
	for i := range shuffledKeys {
		shuffledKeys[i] = Shuffle(sortedKeys[:5])
	}

	for _, keys := range shuffledKeys {
		srr := ArrayToStrings(keys)
		t.Log(srr)

		arr := StringsToArray[[]byte](srr)
		t.Log(ArrayToStrings(arr))

		if !reflect.DeepEqual(keys, arr) {
			t.Fatal("Should be equal")
		}
	}
}

// Test basic character mappings
func TestEncodingDecoding(t *testing.T) {
	tests := []struct {
		input         string
		enc64Expected string
		sec64Expected string
		ascExpected   string
	}{
		{"Hello WORLD!", "helloSworldK", "helloSworld!", "hello world!"}, // A-Z → a-z, space → S
		// {"123", "123", "123"},                               // Numbers remain unchanged
		// {"[test] {code}", "(test)S(code)", "(test) (code)"}, // Brackets → ()
		// {"Hello\nWorld", "helloNworld", "hello\nworld"},     // Newline → N
		// {`=+-*/\%^<>!?@#$&(),;:'"_.`, `=+-*/\%^<>!?@#$&(),;:'"_.`, `=+-*/\%^<>!?@#$&(),;:'"_.`},
		// {"_~", "_Q", "_\x00"}, // `~` is unknown → Q
		// {" ", "S"},                         // Space → S
		// {"\n", "N"},                        // Newline → N
		// {"\x00", "Q"},                      // Null → Q
	}

	for _, tt := range tests {
		encoded := AsciiToEncEncode(tt.input)
		if encoded != tt.sec64Expected {
			t.Errorf("encode(%q) = %q; want %q", tt.input, encoded, tt.sec64Expected)
		}

		decoded := EncToSEC64Decode(encoded)
		if decoded != tt.enc64Expected {
			t.Errorf("decode(%q) = %q; want %q", encoded, decoded, tt.enc64Expected)
		}
	}
}

// // Test round-trip encoding and decoding
// func TestRoundTrip(t *testing.T) {
// 	inputs := []string{
// 		"Hello WORLD!",
// 		"Testing 123",
// 		"[Brackets] and {Braces}",
// 		"\nNewline and space ",
// 		"<>?@#$%^&*()",
// 		"\x00\x01\x02Invalid ASCII",
// 	}

// 	for _, input := range inputs {
// 		encoded := SEC64Encode(input)
// 		decoded := SEC64Decode(encoded)

// 		if decoded != input {
// 			t.Errorf("Round-trip failed: input=%q, encoded=%q, decoded=%q", input, encoded, decoded)
// 		}
// 	}
// }
