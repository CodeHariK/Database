package utils

// Map applies a function to each element in the slice and returns a new slice.
func Map[T, R any](arr []T, fn func(T) R) []R {
	result := make([]R, len(arr)) // Allocate a new slice for results
	for i, v := range arr {
		result[i] = fn(v) // Apply transformation function
	}
	return result
}
