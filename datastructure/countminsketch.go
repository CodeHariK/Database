package datastructure

import (
	"errors"

	"github.com/zeebo/xxh3"
)

type CountMinSketch struct {
	freq   [][]uint64
	hasher []*xxh3.Hasher
	m      uint64
	k      uint64
	count  uint64
}

func NewCountMinSketch(m, k uint64) CountMinSketch {
	hasher := make([]*xxh3.Hasher, k)
	for i := range hasher {
		h := xxh3.NewSeed(uint64(i))
		hasher[i] = h
	}

	freq := make([][]uint64, k)
	for i := uint64(0); i < k; i++ {
		freq[i] = make([]uint64, m) // each row is size m
	}

	return CountMinSketch{
		freq:   freq,
		hasher: hasher,
		m:      m,
		k:      k,
	}
}

func (c *CountMinSketch) Add(data []byte) {
	for i := uint64(0); i < c.k; i++ {
		pos := xxh3.HashSeed(data, i) % c.m
		c.freq[i][pos]++
	}
	c.count++
}

func (c *CountMinSketch) Count(data []byte) uint64 {
	min := uint64(^uint64(0)) // max uint64
	for i := uint64(0); i < c.k; i++ {
		pos := xxh3.HashSeed(data, i) % c.m
		count := c.freq[i][pos]
		if count < min {
			min = count
		}
	}
	return min
}

// Merge combines this CountMinSketch with another. Returns an error if the
// matrix width and depth are not equal.
func (c *CountMinSketch) Merge(other *CountMinSketch) error {
	if c.k != other.k {
		return errors.New("matrix depth must match")
	}

	if c.m != other.m {
		return errors.New("matrix width must match")
	}

	for i := uint64(0); i < c.k; i++ {
		for j := uint64(0); j < c.m; j++ {
			c.freq[i][j] += other.freq[i][j]
		}
	}

	c.count += other.count
	return nil
}
