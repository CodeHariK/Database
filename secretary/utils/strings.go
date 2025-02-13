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
func GenerateRandomString(l int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, l)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

var generateSeq int32 = 0

func GenerateSeqString(length int) string {
	generateSeq += 1
	return fmt.Sprintf("%0*d", length, generateSeq)
}

func GenerateSeqRandomString(length int, pad int, value ...string) string {
	generateSeq += 1
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	s := fmt.Sprintf("%0*d:%s:%s", pad, generateSeq, value, string(b))
	return s[:length]
}

func BytesToStrings(byteSlices [][]byte) []string {
	strs := make([]string, len(byteSlices))
	for i, b := range byteSlices {
		strs[i] = string(b) // Convert []byte to string
	}
	return strs
}

func CompareStringArray(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
