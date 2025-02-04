package utils

import (
	"math/rand"
)

// GenerateRandomSlice creates a slice of the given size and fills it with random values.
func GenerateRandomSlice[T int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](size int) []T {
	slice := make([]T, size)

	for i := range slice {
		switch any(slice[i]).(type) {
		case int8:
			slice[i] = T(rand.Intn(128) - 128) // int8 range: -128 to 127
		case int16:
			slice[i] = T(rand.Intn(32768) - 32768) // int16 range: -32768 to 32767
		case int32:
			slice[i] = T(rand.Int31())
		case int64:
			slice[i] = T(rand.Int63())
		case uint8:
			slice[i] = T(rand.Intn(256)) // uint8 range: 0 to 255
		case uint16:
			slice[i] = T(rand.Intn(65536)) // uint16 range: 0 to 65535
		case uint32:
			slice[i] = T(rand.Uint32())
		case uint64:
			slice[i] = T(rand.Uint64())
		case uint:
			slice[i] = T(rand.Uint32()) // Assuming uint is at least 32 bits
		}
	}

	return slice
}

func MakeByteArray(size int, c rune) []byte {
	b := make([]byte, size)

	for i := range b {
		b[i] = byte(c)
	}

	return b
}
