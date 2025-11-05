package ngram

import (
	"fmt"
	"testing"
)

func TestNGram(t *testing.T) {
	text := "hello world this is a test"

	fmt.Println("Bigrams:", GenerateNGrams(text, 2))
	fmt.Println("Trigrams:", GenerateNGrams(text, 3))
	fmt.Println("Four-grams:", GenerateNGrams(text, 4))
}
