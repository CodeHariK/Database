package utils

import (
	"math/rand"
)

func Ternary[T any](condition bool, value1, value2 T) T {
	if condition {
		return value1
	}
	return value2
}

func Shuffle[T any](arr []T) []T {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr
}
