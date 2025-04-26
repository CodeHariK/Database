package datastructure

import "github.com/zeebo/xxh3"

type HyperLogLog struct {
	registers []uint8 // Buckets: each register holds the max number of leading zeros
	p         uint8   // Precision: number of bits used for bucket index
}

func NewHyperLogLog(p uint8) *HyperLogLog {
	m := 1 << p // Number of registers
	return &HyperLogLog{
		registers: make([]uint8, m),
		p:         p,
	}
}

func (hll *HyperLogLog) Add(data []byte) {
	hash := xxh3.Hash(data) // Get 64-bit hash

	p := hll.p
	idx := hash >> (64 - p) // Take the top p bits as index
	w := hash << p          // Shift left, removing top p bits

	// Count leading zeros in w
	var rank uint8 = 1
	for w&(1<<(63)) == 0 && rank <= 64-p {
		rank++
		w <<= 1
	}

	// Update the register
	if rank > hll.registers[idx] {
		hll.registers[idx] = rank
	}
}
