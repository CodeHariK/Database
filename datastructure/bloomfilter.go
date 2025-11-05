package datastructure

import (
	"errors"

	"github.com/zeebo/xxh3"
)

type BloomFilter struct {
	bitset []byte
	hasher []*xxh3.Hasher
	m      uint64
	k      uint64
}

func (b *BloomFilter) setBit(pos uint64) {
	b.bitset[pos/8] |= 1 << (pos % 8)
}

func (b *BloomFilter) getBit(pos uint64) bool {
	return b.bitset[pos/8]&(1<<(pos%8)) != 0
}

func (b *BloomFilter) Add(data []byte) {
	for i := 0; i < int(b.k); i++ {
		h := b.hasher[i]
		h.Reset()
		h.Write(data)
		hash := h.Sum64()
		b.setBit(hash % b.m)
	}
}

func (b *BloomFilter) Check(data []byte) bool {
	for i := 0; i < int(b.k); i++ {
		h := b.hasher[i]
		h.Reset()
		h.Write(data)
		hash := h.Sum64()
		if !b.getBit(hash % b.m) {
			return false
		}
	}
	return true
}

func (b *BloomFilter) Union(a *BloomFilter) (err error) {
	if b.m != a.m {
		return errors.New("the bloom filters have the different sizes")
	}

	if b.k != a.k {
		return errors.New("the bloom filters have the different number of hash functions")
	}

	for i := uint64(0); i < b.m; i++ {
		if b.getBit(i) || a.getBit(i) {
			b.setBit(i)
		}
	}

	return nil
}

func NewBloomFilter(m uint64, k uint64) BloomFilter {
	hasher := make([]*xxh3.Hasher, k)
	for i := range hasher {
		h := xxh3.NewSeed(uint64(i))
		hasher[i] = h
	}
	return BloomFilter{
		bitset: make([]byte, (m+7)/8),
		hasher: hasher,
		m:      m,
		k:      k,
	}
}
