package ngram

import (
	"strings"

	"github.com/codeharik/secretary/utils"
)

// GenerateNGrams creates n-grams of the given size from a string.
func GenerateNGrams(text string, n int) []string {
	var ngrams []string
	words := strings.Fields(text) // Split into words

	utils.Log(words)

	if len(words) < n {
		return ngrams // Not enough words to form an n-gram
	}

	for i := 0; i <= len(words)-n; i++ {
		ngrams = append(ngrams, strings.Join(words[i:i+n], " "))
		utils.Log(ngrams)
	}

	return ngrams
}
