package fbptree

import "bytes"

func compare(x, y []byte) int {
	return bytes.Compare(x, y)
}

func less(x, y []byte) bool {
	return compare(x, y) < 0
}

func copyBytes(s []byte) []byte {
	c := make([]byte, len(s))
	copy(c, s)

	return c
}

func ceil(x, y int) int {
	d := (x / y)
	if x%y == 0 {
		return d
	}

	return d + 1
}
