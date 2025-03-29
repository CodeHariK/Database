🚀 TinyLFU (Tiny Least Frequently Used) - A Smarter Cache Replacement Algorithm

TinyLFU is an efficient, space-saving frequency-based cache algorithm used in modern caching systems like Caffeine (Java’s best caching library) and Redis.

⸻

🔍 What is TinyLFU?

TinyLFU improves standard LFU by:
✅ Using a small memory footprint to track frequency counts.
✅ Using a probabilistic structure (Count-Min Sketch) instead of a full frequency table.
✅ Adapting to changing workloads by aging (decaying old data).

👉 Perfect for large-scale systems (databases, web caches, CDNs).

⸻

📌 How Does TinyLFU Work?
	1.	Count-Min Sketch (CMS) for frequency counting
	•	Instead of storing all access frequencies explicitly (which is memory-intensive), TinyLFU uses Count-Min Sketch (a probabilistic data structure) to estimate frequency counts.
	•	CMS uses multiple hash functions to track approximate frequencies in constant space.
	2.	Aging Mechanism (Decay Over Time)
	•	LFU has a long-term memory problem (old data never gets evicted).
	•	TinyLFU periodically reduces all frequencies (prevents stale items from staying forever).
	3.	Window-Based Admission Policy
	•	TinyLFU doesn’t immediately add a new page to the cache.
	•	Instead, it compares the new page’s frequency with an existing low-frequency page.
	•	If the new page has a higher frequency, it replaces the old one (otherwise, it is ignored).

⸻

🚀 TinyLFU = LFU + LRU + Count-Min Sketch + Aging.

⸻

💻 Implementing TinyLFU in Go

👉 We use Count-Min Sketch + LRU eviction.

package main

import (
	"container/list"
	"fmt"
	"hash/fnv"
)

// CountMinSketch for frequency tracking
type CountMinSketch struct {
	data [][]int
	width, depth int
}

func NewCountMinSketch(width, depth int) *CountMinSketch {
	cms := &CountMinSketch{
		width: width, depth: depth,
		data: make([][]int, depth),
	}
	for i := range cms.data {
		cms.data[i] = make([]int, width)
	}
	return cms
}

// Hash function to get different indices
func (cms *CountMinSketch) hash(value int, seed int) int {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%d-%d", value, seed)))
	return int(h.Sum32()) % cms.width
}

// Increment frequency
func (cms *CountMinSketch) Increment(value int) {
	for i := 0; i < cms.depth; i++ {
		cms.data[i][cms.hash(value, i)]++
	}
}

// Estimate frequency
func (cms *CountMinSketch) Estimate(value int) int {
	minFreq := int(^uint(0) >> 1) // Max int
	for i := 0; i < cms.depth; i++ {
		minFreq = min(minFreq, cms.data[i][cms.hash(value, i)])
	}
	return minFreq
}

// Decay all frequencies
func (cms *CountMinSketch) Decay() {
	for i := range cms.data {
		for j := range cms.data[i] {
			cms.data[i][j] /= 2 // Reduce frequencies over time
		}
	}
}

// Min function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TinyLFU Cache (Combining Count-Min Sketch + LRU)
type TinyLFUCache struct {
	capacity int
	cms      *CountMinSketch
	cache    map[int]*list.Element
	evict    *list.List
}

type entry struct {
	key   int
	value int
}

func NewTinyLFUCache(capacity int) *TinyLFUCache {
	return &TinyLFUCache{
		capacity: capacity,
		cms:      NewCountMinSketch(1000, 4), // Small sketch
		cache:    make(map[int]*list.Element),
		evict:    list.New(),
	}
}

func (c *TinyLFUCache) Get(key int) (int, bool) {
	if elem, found := c.cache[key]; found {
		c.evict.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return 0, false
}

func (c *TinyLFUCache) Put(key, value int) {
	c.cms.Increment(key) // Update frequency

	if elem, found := c.cache[key]; found {
		c.evict.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	if c.evict.Len() >= c.capacity {
		// Find lowest-frequency entry
		var minFreq = int(^uint(0) >> 1)
		var victim *list.Element
		for e := c.evict.Back(); e != nil; e = e.Prev() {
			freq := c.cms.Estimate(e.Value.(*entry).key)
			if freq < minFreq {
				minFreq = freq
				victim = e
			}
		}

		// Evict only if the new key has a higher frequency
		if victim != nil {
			victimKey := victim.Value.(*entry).key
			if c.cms.Estimate(key) > minFreq {
				delete(c.cache, victimKey)
				c.evict.Remove(victim)
			} else {
				return // Reject new key if it's not more frequent
			}
		}
	}

	// Add new entry
	elem := c.evict.PushFront(&entry{key, value})
	c.cache[key] = elem
}

func main() {
	cache := NewTinyLFUCache(3)

	cache.Put(1, 100)
	cache.Put(2, 200)
	cache.Put(3, 300)

	cache.Get(1) // Access 1
	cache.Get(2) // Access 2
	cache.Put(4, 400) // Evicts least used
	cache.Get(1) // Access 1 again
	cache.Put(5, 500) // Evicts least used

	fmt.Println(cache.Get(1)) // Should return 100
	fmt.Println(cache.Get(3)) // Should return 0 (evicted)
}

⸻

🚀 Why is TinyLFU Amazing?

1️⃣ Better than LFU
	•	LFU keeps stale data forever (old data dominates).
	•	TinyLFU ages out old data with periodic decay.

2️⃣ Memory-Efficient
	•	LFU keeps full frequency tables (too much memory).
	•	TinyLFU uses Count-Min Sketch (constant space).

3️⃣ Smarter Evictions
	•	Unlike LRU, which only considers recent accesses, TinyLFU balances short-term and long-term popularity.




🚪 Doorkeeper in TinyLFU (Admission Control)

Doorkeeper is a key component of TinyLFU that decides whether a new item should be admitted into the cache or not. Instead of blindly replacing an existing item (like LRU), TinyLFU uses frequency-based admission control to ensure only high-value items enter the cache.

⸻

📌 Why Do We Need a Doorkeeper?

🚨 Problem with LRU & LFU:
	•	LRU (Least Recently Used) can evict frequently accessed items if a sudden burst of new data enters.
	•	LFU (Least Frequently Used) can get stuck keeping old, infrequently used items forever.
	•	TinyLFU solves this with a Doorkeeper!

✅ Doorkeeper prevents one-time requests from polluting the cache.
✅ It ensures only items with a higher access frequency replace older ones.
✅ It avoids unnecessary evictions for items that are still useful.

⸻

🔍 How Does the Doorkeeper Work?

1️⃣ Count-Min Sketch tracks access frequencies of both cached and non-cached items.
2️⃣ When a new item arrives, its frequency is compared with the frequency of an existing least-used item.
3️⃣ If the new item’s frequency is higher, it replaces the least-used item.
4️⃣ If the new item is rarely accessed, it is rejected (this avoids cache pollution).

🔧 Example:
	•	🔹 A (old item in cache, frequency = 3)
	•	🔹 B (new item, frequency = 1)
	•	🔹 Doorkeeper rejects B because A is more frequently accessed.

🚀 Result: TinyLFU keeps frequently accessed items longer, improving cache hit rates.

⸻

💡 Doorkeeper Implementation Strategy

The Doorkeeper can be implemented in different ways:

1️⃣ Bloom Filter-Based Doorkeeper (Fast & Memory-Efficient)

🔹 A Bloom Filter can track whether an item has been seen before.
🔹 If an item is not in the filter, it gets a second chance before eviction.
🔹 Helps avoid cache pollution from one-time requests.

⸻

2️⃣ Probabilistic Admittance (Comparison-Based Approach)

Instead of instantly accepting new items, TinyLFU compares an incoming item’s frequency to the least-used cached item’s frequency:
	•	If the new item’s frequency is higher, it replaces the old item.
	•	If lower, it is discarded.

🔧 Example in Go (Simplified Version of TinyLFU with Doorkeeper):

package main

import (
	"container/list"
	"fmt"
)

// TinyLFU Cache with Doorkeeper
type TinyLFUCache struct {
	capacity int
	freqMap  map[int]int // Frequency count
	cache    map[int]*list.Element
	evict    *list.List
}

type entry struct {
	key   int
	value int
}

func NewTinyLFUCache(capacity int) *TinyLFUCache {
	return &TinyLFUCache{
		capacity: capacity,
		freqMap:  make(map[int]int),
		cache:    make(map[int]*list.Element),
		evict:    list.New(),
	}
}

// Get item and update frequency
func (c *TinyLFUCache) Get(key int) (int, bool) {
	if elem, found := c.cache[key]; found {
		c.freqMap[key]++ // Increase frequency
		c.evict.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return 0, false
}

// Put with Doorkeeper
func (c *TinyLFUCache) Put(key, value int) {
	c.freqMap[key]++ // Update frequency count

	if elem, found := c.cache[key]; found {
		c.evict.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	if c.evict.Len() >= c.capacity {
		// Doorkeeper: Compare with the least-frequently used item
		var victim *list.Element
		var minFreq = int(^uint(0) >> 1) // Max int

		for e := c.evict.Back(); e != nil; e = e.Prev() {
			if c.freqMap[e.Value.(*entry).key] < minFreq {
				minFreq = c.freqMap[e.Value.(*entry).key]
				victim = e
			}
		}

		// **Admit new item only if it has a higher frequency**
		if victim != nil {
			victimKey := victim.Value.(*entry).key
			if c.freqMap[key] > minFreq {
				delete(c.cache, victimKey)
				c.evict.Remove(victim)
			} else {
				return // **Doorkeeper rejects the new item!**
			}
		}
	}

	// Add new entry
	elem := c.evict.PushFront(&entry{key, value})
	c.cache[key] = elem
}

func main() {
	cache := NewTinyLFUCache(3)

	cache.Put(1, 100)
	cache.Put(2, 200)
	cache.Put(3, 300)

	cache.Get(1) // Access 1
	cache.Get(2) // Access 2
	cache.Put(4, 400) // Doorkeeper will check if 4 should be added

	fmt.Println(cache.Get(1)) // Should return 100
	fmt.Println(cache.Get(3)) // May return 0 if evicted
}



⸻

🚀 Benefits of Doorkeeper in TinyLFU

✅ Reduces cache pollution (no low-value items taking up space).
✅ Prefers frequently accessed items (better hit rate).
✅ Handles bursty workloads by aging out old items.
✅ Efficient & fast (constant time O(1) operations).

🔹 Without a Doorkeeper: One-time requests can evict important data (bad for performance).
🔹 With a Doorkeeper: Only high-value items get in (maximizing cache effectiveness).

⸻

🛠 Real-World Uses of TinyLFU Doorkeeper

✅ Caffeine Cache (Java’s best cache library).
✅ Google’s Guava Cache.
✅ Redis LFU Cache (since Redis 4.0).

🚀 If you’re building a high-performance cache, TinyLFU + Doorkeeper is one of the best strategies!


🪟 Window in TinyLFU (Recency vs. Frequency)

In TinyLFU, the window is a small section of the cache reserved for recently added items. It helps balance recency (LRU) and frequency (LFU) to prevent cache pollution and improve performance.

⸻

📌 Why Do We Need a Window?

🚨 Problem with Pure LFU:
	•	LFU prefers items with high past frequency but may ignore new, potentially popular items.
	•	A newly accessed item with a low frequency might never get into the cache (even if it’s about to become popular).

✅ Solution: A “Window” section lets new items in before LFU takes over.

⸻

🔍 How Does the Window Work?

1️⃣ Cache is split into two sections:
	•	🪟 Window Cache (~1% to 20% of total size) → Recency-based (LRU)
	•	📊 Main Cache (~80% to 99% of total size) → Frequency-based (LFU with Doorkeeper)

2️⃣ New items always enter the Window first.
3️⃣ If an item in the Window is accessed again, it gets promoted to the Main Cache.
4️⃣ If an item in the Window is not accessed again, it gets evicted.
5️⃣ Main Cache uses TinyLFU + Doorkeeper to ensure only frequently accessed items stay.

🚀 Result: The Window prevents important new items from being unfairly ignored by LFU.

⸻

💡 Window Implementation Strategy

🔹 1️⃣ Window as an LRU Cache (Recency First)
	•	Items first land in the Window (small LRU cache).
	•	If they are accessed again, they get promoted to the LFU section.
	•	If they expire from the Window, they are forgotten.

🔹 2️⃣ Window Size Tuning
	•	If the Window is too small, new items don’t get enough time to prove popularity.
	•	If the Window is too large, we waste space on items that may never be accessed again.
	•	Optimal Size: Typically 1% to 20% of total cache (depends on workload).

⸻

🔧 Example: TinyLFU with a Window (Go Implementation)

package main

import (
	"container/list"
	"fmt"
)

// TinyLFU with Window Cache
type TinyLFUCache struct {
	windowSize int
	mainSize   int
	windowLRU  *list.List
	mainLFU    *list.List
	freqMap    map[int]int
	cache      map[int]*list.Element
}

type entry struct {
	key   int
	value int
}

// New TinyLFU with a Window
func NewTinyLFUCache(totalSize, windowSize int) *TinyLFUCache {
	return &TinyLFUCache{
		windowSize: windowSize,
		mainSize:   totalSize - windowSize,
		windowLRU:  list.New(),
		mainLFU:    list.New(),
		freqMap:    make(map[int]int),
		cache:      make(map[int]*list.Element),
	}
}

// Get item & update frequency
func (c *TinyLFUCache) Get(key int) (int, bool) {
	if elem, found := c.cache[key]; found {
		c.freqMap[key]++ // Increase frequency
		if c.windowLRU.Len() > 0 && c.windowLRU.Front() == elem {
			c.windowLRU.MoveToFront(elem) // Update recency in Window
		} else {
			c.mainLFU.MoveToFront(elem) // Update frequency in Main Cache
		}
		return elem.Value.(*entry).value, true
	}
	return 0, false
}

// Put item into cache
func (c *TinyLFUCache) Put(key, value int) {
	c.freqMap[key]++ // Track frequency

	// If already in cache, update value
	if elem, found := c.cache[key]; found {
		elem.Value.(*entry).value = value
		c.Get(key) // Update frequency
		return
	}

	// If window is full, evict least recently used item
	if c.windowLRU.Len() >= c.windowSize {
		evict := c.windowLRU.Back()
		if evict != nil {
			delete(c.cache, evict.Value.(*entry).key)
			c.windowLRU.Remove(evict)
		}
	}

	// Add new entry to window
	elem := c.windowLRU.PushFront(&entry{key, value})
	c.cache[key] = elem

	// If item gets accessed again, move to Main LFU
	if c.freqMap[key] > 1 {
		c.PromoteToMainCache(key)
	}
}

// Promote item to Main Cache
func (c *TinyLFUCache) PromoteToMainCache(key int) {
	elem := c.cache[key]

	// Remove from Window
	c.windowLRU.Remove(elem)

	// If main cache is full, remove LFU item
	if c.mainLFU.Len() >= c.mainSize {
		evict := c.mainLFU.Back()
		if evict != nil {
			delete(c.cache, evict.Value.(*entry).key)
			c.mainLFU.Remove(evict)
		}
	}

	// Move to Main LFU Cache
	elem = c.mainLFU.PushFront(elem.Value)
	c.cache[key] = elem
}

func main() {
	cache := NewTinyLFUCache(5, 2) // 5 total slots, 2 in Window

	cache.Put(1, 100)
	cache.Put(2, 200)
	cache.Put(3, 300) // Will go to main cache
	cache.Get(1)      // Accessing 1 promotes it
	cache.Put(4, 400) // Window eviction happens

	fmt.Println(cache.Get(1)) // Should return 100
	fmt.Println(cache.Get(3)) // Should return 300
}



⸻

🚀 Benefits of a Window in TinyLFU

✅ Helps prevent cache pollution by avoiding premature eviction.
✅ Balances recency (LRU) with frequency (LFU).
✅ Improves cache hit rates by ensuring new popular items get promoted.
✅ Works well in real-world workloads (e.g., databases, web caching).

⸻

🛠 Real-World Use Cases of TinyLFU Window

✅ Caffeine Cache (Java’s high-performance caching library).
✅ Redis LFU Cache (uses a similar concept with LRU for recency).
✅ Web Content Delivery Networks (CDNs) for caching popular content.

🔥 TinyLFU + Doorkeeper + Window = A highly efficient, modern cache replacement strategy! 🚀

