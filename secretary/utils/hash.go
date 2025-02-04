package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

// Function to compute MD5 hash of a struct
func Md5Struct(data interface{}) (string, error) {
	// Serialize the struct to JSON
	serialized, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Compute MD5 hash of the serialized data
	hash := md5.New()
	hash.Write(serialized)

	// Get the hash sum as a byte slice
	hashBytes := hash.Sum(nil)

	// Return the hash as a hexadecimal string
	return hex.EncodeToString(hashBytes), nil
}
