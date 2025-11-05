package datastructure

import (
	"math"

	"github.com/zeebo/xxh3"
)

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

func alpha(m int) float64 {
	switch m {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	default:
		return 0.7213 / (1 + 1.079/float64(m))
	}
}

func (hll *HyperLogLog) Estimate() float64 {
	m := float64(len(hll.registers))
	sum := 0.0
	zeros := 0

	for _, reg := range hll.registers {
		sum += 1.0 / float64(uint64(1)<<reg) // 2^-reg
		if reg == 0 {
			zeros++
		}
	}

	alphaM := alpha(int(m))
	est := alphaM * m * m / sum

	// 🔥 Small range correction
	if est <= 5.0*m {
		if zeros != 0 {
			est = m * math.Log(m/float64(zeros))
		}
	}

	return est
}
