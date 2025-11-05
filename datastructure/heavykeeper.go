package datastructure

import (
	"container/heap"
	"math"
	"math/rand/v2"

	"github.com/zeebo/xxh3"
)

// FlowCount holds the flow name and its count
type FlowCount struct {
	Flow  string
	Count uint32
}

type bucket struct {
	fingerprint uint32
	count       uint32
}

type HeavyKeeper struct {
	decay   float64
	buckets [][]bucket
	heap    minHeap
}

func New(k int, decay float64) *HeavyKeeper {
	if k < 1 {
		panic("k must be >= 1")
	}
	if decay <= 0 || decay > 1 {
		panic("decay must be between (0,1]")
	}

	width := int(float64(k) * math.Log(float64(k)))
	if width < 256 {
		width = 256
	}

	depth := int(math.Log(float64(k)))
	if depth < 3 {
		depth = 3
	}

	buckets := make([][]bucket, depth)
	for i := range buckets {
		buckets[i] = make([]bucket, width)
	}

	return &HeavyKeeper{
		decay:   decay,
		buckets: buckets,
		heap:    make(minHeap, k),
	}
}

func fingerprint(flow string) uint32 {
	return uint32(xxh3.Hash([]byte(flow)))
}

func slot(flow string, row uint32, width uint32) uint32 {
	data := []byte(flow)
	seed := uint64(row) // use row index as seed
	h := xxh3.NewSeed(seed)
	h.Write(data)
	return uint32(h.Sum64() % uint64(width))
}

func (hk *HeavyKeeper) Sample(flow string, incr uint32) {
	fp := fingerprint(flow)
	var maxCount uint32

	// Process each row in the bucket
	for i, row := range hk.buckets {
		j := slot(flow, uint32(i), uint32(len(row)))

		// Update bucket counts
		if row[j].count == 0 {
			row[j].fingerprint = fp
			row[j].count = incr
		} else if row[j].fingerprint == fp {
			row[j].count += incr
		} else {
			// Decay old flow and update with new fingerprint
			for decays := uint32(0); decays < incr; decays++ {
				chance := math.Pow(hk.decay, float64(row[j].count))
				if rand.Float64() < chance {
					row[j].count--
					if row[j].count == 0 {
						row[j].fingerprint = fp
						row[j].count = 1
						break
					}
				}
			}
		}
		maxCount = max(maxCount, row[j].count)
	}

	// Now we update the heap with the new count.
	// If the flow's count is greater than or equal to the minimum count in the heap, we update it.

	// Check if the flow is in the heap
	i := hk.heap.Find(flow)
	if i > -1 {
		// Flow is in heap, update its count
		hk.heap[i].Count = maxCount
		heap.Fix(&hk.heap, i) // Reorder the heap to maintain the min-heap property
	} else {
		// Flow is not in the heap, check if we need to insert it
		heapMin := hk.heap.Min()
		if maxCount >= heapMin {
			hk.heap[0].Flow = flow
			hk.heap[0].Count = maxCount
			heap.Fix(&hk.heap, 0) // Reorder the heap to maintain the min-heap property
		}
	}
}

// minHeap is a custom type that implements the heap.Interface for min-heap behavior
type minHeap []FlowCount

// Implement heap.Interface methods for minHeap

// Len returns the number of elements in the heap
func (h minHeap) Len() int {
	return len(h)
}

// Less reports whether the element with index i should sort before the element with index j
func (h minHeap) Less(i, j int) bool {
	return h[i].Count < h[j].Count // Min-heap: smaller count comes first
}

// Swap swaps the elements with indexes i and j
func (h minHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push adds an element to the heap
func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(FlowCount))
}

// Pop removes and returns the smallest element from the heap
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Min returns the smallest count in the heap (top element)
func (h minHeap) Min() uint32 {
	return h[0].Count
}

// Find finds the index of the flow in the heap, -1 if not found
func (h minHeap) Find(flow string) int {
	for i := range h {
		if h[i].Flow == flow {
			return i
		}
	}
	return -1
}
