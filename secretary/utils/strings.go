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
