package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"sync/atomic"
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

func GenerateSeqString(sequence *uint64, length int, increment uint64) string {
	atomic.AddUint64(sequence, increment)
	key := fmt.Sprintf("%0*d", length, *sequence)
	return key
}

func GenerateSeqRandomString(sequence *uint64, length int, increment uint64, pad int, value ...string) string {
	atomic.AddUint64(sequence, increment)
	str := fmt.Sprintf("%0*d:%s:%s", pad, *sequence, value, GenerateRandomString(length))
	return str[:length]
}

func ArrayContains[T string](arr []T, target T) bool {
	for _, arg := range arr {
		if arg == target {
			return true
		}
	}
	return false
}
