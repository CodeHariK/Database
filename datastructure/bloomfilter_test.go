package datastructure

import (
	"testing"
)

func TestBloomFilter(t *testing.T) {
	bf := NewBloomFilter(1024, 3) // 1024 bits, 3 hash functions

	wordsToAdd := [][]byte{
		[]byte("hello"),
		[]byte("world"),
		[]byte("golang"),
	}

	wordsNotAdded := [][]byte{
		[]byte("python"),
		[]byte("java"),
		[]byte("rust"),
	}

	// Add words
	for _, word := range wordsToAdd {
		bf.Add(word)
	}

	// Check words that were added
	for _, word := range wordsToAdd {
		if !bf.Check(word) {
			t.Errorf("Expected word %q to be found in BloomFilter", word)
		}
	}

	// Check words that were NOT added
	falsePositiveCount := 0
	for _, word := range wordsNotAdded {
		if bf.Check(word) {
			falsePositiveCount++
		}
	}

	// Allow a small false positive rate
	if falsePositiveCount > 1 {
		t.Errorf("Too many false positives: got %d, want <= 1", falsePositiveCount)
	}
}

func BenchmarkBloomFilterAdd(b *testing.B) {
	bf := NewBloomFilter(1<<20, 5) // 1 million bits, 5 hash functions
	data := []byte("benchmark-data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(data)
	}
}

func BenchmarkBloomFilterCheck(b *testing.B) {
	bf := NewBloomFilter(1<<20, 5)
	data := []byte("benchmark-data")
	bf.Add(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Check(data)
	}
}
