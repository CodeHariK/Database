package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	MaxLevel  = 16   // Max levels in the skip list
	P         = 0.25 // Probability for level increase
	LoadLimit = 4    // Max items per shard before resizing
)

type Node struct {
	key   string
	value int
	next  [MaxLevel]unsafe.Pointer
	level int
}

type Shard struct {
	head unsafe.Pointer // Pointer to skip list head
	size int64          // Atomic size counter
}

type ShardedMap struct {
	shards   []*Shard
	shardCnt int64
}

func NewShardedMap() *ShardedMap {
	shardCount := runtime.NumCPU() * 2 // Dynamically scale with CPU cores
	shards := make([]*Shard, shardCount)
	for i := range shards {
		shards[i] = &Shard{head: unsafe.Pointer(newNode("", 0, MaxLevel))}
	}
	return &ShardedMap{shards: shards, shardCnt: int64(shardCount)}
}

// Generates a random level for the skip list node
func randomLevel() int {
	level := 1
	for rand.Float32() < P && level < MaxLevel {
		level++
	}
	return level
}

// Creates a new node
func newNode(key string, value int, level int) *Node {
	return &Node{key: key, value: value, level: level}
}

// Hash function to distribute keys across shards
func hash(key string) int {
	hash := 0
	for i := 0; i < len(key); i++ {
		hash = (hash*31 + int(key[i])) % 99991 // Large prime for distribution
	}
	return hash
}

// GetShard selects the shard
func (m *ShardedMap) GetShard(key string) *Shard {
	index := hash(key) % len(m.shards)
	return m.shards[index]
}

// Insert a key-value pair in a lock-free way
func (m *ShardedMap) Put(key string, value int) {
	shard := m.GetShard(key)

	newNode := newNode(key, value, randomLevel())

	for {
		prev := (*Node)(atomic.LoadPointer(&shard.head))
		curr := prev

		// Traverse the skip list and find insertion point
		update := [MaxLevel]*Node{}
		for i := MaxLevel - 1; i >= 0; i-- {
			for (*Node)(atomic.LoadPointer(&curr.next[i])) != nil &&
				(*Node)(atomic.LoadPointer(&curr.next[i])).key < key {
				curr = (*Node)(atomic.LoadPointer(&curr.next[i]))
			}
			update[i] = curr
		}

		// Atomic insert
		next := (*Node)(atomic.LoadPointer(&update[0].next[0]))
		newNode.next[0] = unsafe.Pointer(next)
		if atomic.CompareAndSwapPointer(&update[0].next[0], unsafe.Pointer(next), unsafe.Pointer(newNode)) {
			atomic.AddInt64(&shard.size, 1)
			break
		}
	}

	// Resize if needed
	if atomic.LoadInt64(&shard.size) > LoadLimit {
		m.Resize()
	}
}

// Retrieve a value from the skip list
func (m *ShardedMap) Get(key string) (int, bool) {
	shard := m.GetShard(key)
	curr := (*Node)(atomic.LoadPointer(&shard.head))

	for i := MaxLevel - 1; i >= 0; i-- {
		for (*Node)(atomic.LoadPointer(&curr.next[i])) != nil &&
			(*Node)(atomic.LoadPointer(&curr.next[i])).key < key {
			curr = (*Node)(atomic.LoadPointer(&curr.next[i]))
		}
	}

	curr = (*Node)(atomic.LoadPointer(&curr.next[0]))

	if curr != nil && curr.key == key {
		return curr.value, true
	}
	return 0, false
}

// Delete a key from the skip list
func (m *ShardedMap) Delete(key string) {
	shard := m.GetShard(key)

	for {
		prev := (*Node)(atomic.LoadPointer(&shard.head))
		curr := prev
		update := [MaxLevel]*Node{}

		// Traverse the skip list and find deletion point
		for i := MaxLevel - 1; i >= 0; i-- {
			for (*Node)(atomic.LoadPointer(&curr.next[i])) != nil &&
				(*Node)(atomic.LoadPointer(&curr.next[i])).key < key {
				curr = (*Node)(atomic.LoadPointer(&curr.next[i]))
			}
			update[i] = curr
		}

		curr = (*Node)(atomic.LoadPointer(&update[0].next[0]))

		// If key does not exist, return
		if curr == nil || curr.key != key {
			return
		}

		// Atomic removal
		for i := 0; i < curr.level; i++ {
			next := (*Node)(atomic.LoadPointer(&curr.next[i]))
			atomic.CompareAndSwapPointer(&update[i].next[i], unsafe.Pointer(curr), unsafe.Pointer(next))
		}

		atomic.AddInt64(&shard.size, -1)
		break
	}
}

// Resize function (Sharding Expansion)
func (m *ShardedMap) Resize() {
	newShardCount := atomic.LoadInt64(&m.shardCnt) * 2
	newShards := make([]*Shard, newShardCount)

	for i := range newShards {
		newShards[i] = &Shard{head: unsafe.Pointer(newNode("", 0, MaxLevel))}
	}

	// Rehash existing keys
	for _, shard := range m.shards {
		curr := (*Node)(atomic.LoadPointer(&shard.head))
		for curr != nil {
			if curr.key != "" {
				index := hash(curr.key) % int(newShardCount)
				newShards[index].head = shard.head
			}
			curr = (*Node)(atomic.LoadPointer(&curr.next[0]))
		}
	}

	m.shards = newShards
	atomic.StoreInt64(&m.shardCnt, newShardCount)
}

func main() {
	m := NewShardedMap()

	m.Put("Alice", 100)
	m.Put("Bob", 200)
	m.Put("Charlie", 300)

	val, found := m.Get("Alice")
	fmt.Println("Alice:", val, "Found:", found)

	m.Delete("Alice")

	val, found = m.Get("Alice")
	fmt.Println("Alice after delete:", val, "Found:", found)

	// Simulate high concurrency
	for i := 0; i < 10000; i++ {
		go m.Put(fmt.Sprintf("User%d", i), i)
	}

	time.Sleep(2 * time.Second) // Let goroutines run
}
