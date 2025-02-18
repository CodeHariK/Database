package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
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

func CompareArray[T any](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func ArrayToStrings[T any](byteSlices []T) []string {
	strs := make([]string, len(byteSlices))
	for i, b := range byteSlices {
		switch v := any(b).(type) {
		case []byte:
			strs[i] = string(v)
		case error:
			strs[i] = v.Error()
		default:
			strs[i] = fmt.Sprint(v)
		}
	}
	return strs
}

// Convert []string back to []T
func StringsToArray[T any](strs []string) []T {
	arr := make([]T, len(strs))
	for i, s := range strs {
		var val any
		var zero T

		// Type switch based on T
		switch any(zero).(type) {
		case []byte:
			val = []byte(s)
		case error:
			val = errors.New(s)
		default:
			val = s // Default to string
		}

		// Type assertion to T
		arr[i] = val.(T)
	}
	return arr
}
