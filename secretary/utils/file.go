package utils

import (
	"fmt"
	"os"
)

// Ensure directory exists
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0o755) // Create directory and parents if needed
		if err != nil {
			return fmt.Errorf("%s %v", path, err)
		}
	}
	return nil
}
