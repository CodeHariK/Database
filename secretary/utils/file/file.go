package file

import (
	"fmt"
	"os"
	"path/filepath"
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

// EnsureFile ensures a file exists, creating parent directories if needed.
func EnsureFile(path string) error {
	dir := filepath.Dir(path)

	// Ensure parent directory exists
	if err := EnsureDir(dir); err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create empty file
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", path, err)
		}
		file.Close()
	}

	return nil
}
