package utils

import (
	"regexp"
	"sync"
)

// Global cache for compiled regex
var rEGEXCache sync.Map

// Get or compile a regex
func GetCompiledRegex(pattern string) (*regexp.Regexp, error) {
	// Check if the regex is already cached
	if r, ok := rEGEXCache.Load(pattern); ok {
		return r.(*regexp.Regexp), nil
	}

	// Compile and store in cache
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	rEGEXCache.Store(pattern, r)
	return r, nil
}

func ValidateRegex(value string, pattern string) bool {
	re, err := GetCompiledRegex(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(value)
}
