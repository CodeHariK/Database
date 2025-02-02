package utils

import "regexp"

// RemoveNonAlphanumeric removes all characters except letters and digits
func RemoveNonAlphanumeric(input string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(input, "")
}
