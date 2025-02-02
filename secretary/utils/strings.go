package utils

import (
	"regexp"
)

var SAFE_COLLECTION_REGEX = regexp.MustCompile(`[^a-zA-Z0-9._]`)

// SafeCollectionString removes all characters except letters and digits, lowercase
func SafeCollectionString(input string) string {
	return SAFE_COLLECTION_REGEX.ReplaceAllString(input, "")
}
