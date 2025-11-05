package main

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

const NumShards = 16 // Number of shards

type Node struct {
	key   string
	value int
	next  unsafe.Pointer
}

type Shard struct {
	head unsafe.Pointer // Pointer to first node
}

type ShardedMap struct {
	shards [NumShards]Shard
}

// Hash function to determine the shard index
func hash(key string) int {
	hash := 0
	for i := 0; i < len(key); i++ {
		hash = (hash*31 + int(key[i])) % NumShards
	}
	return hash
}

// Put inserts or updates a key-value pair
func (m *ShardedMap) Put(key string, value int) {
	index := hash(key)
	shard := &m.shards[index]

	newNode := &Node{key: key, value: value}

	for {
		oldHead := (*Node)(atomic.LoadPointer(&shard.head))
		newNode.next = unsafe.Pointer(oldHead)

		if atomic.CompareAndSwapPointer(&shard.head, unsafe.Pointer(oldHead), unsafe.Pointer(newNode)) {
			return
		}
	}
}

// Get retrieves a value by key
func (m *ShardedMap) Get(key string) (int, bool) {
	index := hash(key)
	shard := &m.shards[index]

	for node := (*Node)(atomic.LoadPointer(&shard.head)); node != nil; node = (*Node)(atomic.LoadPointer(&node.next)) {
		if node.key == key {
			return node.value, true
		}
	}
	return 0, false
}

// Delete removes a key from the map
func (m *ShardedMap) Delete(key string) {
	index := hash(key)
	shard := &m.shards[index]

	for {
		oldHead := (*Node)(atomic.LoadPointer(&shard.head))
		var prev *Node
		curr := oldHead

		// Find the node to delete
		for curr != nil && curr.key != key {
			prev = curr
			curr = (*Node)(atomic.LoadPointer(&curr.next))
		}

		// If key is not found, exit
		if curr == nil {
			return
		}

		// Update the next pointer
		next := (*Node)(atomic.LoadPointer(&curr.next))
		if prev == nil {
			// Delete the head node
			if atomic.CompareAndSwapPointer(&shard.head, unsafe.Pointer(oldHead), unsafe.Pointer(next)) {
				return
			}
		} else {
			// Delete a middle or tail node
			if atomic.CompareAndSwapPointer(&prev.next, unsafe.Pointer(curr), unsafe.Pointer(next)) {
				return
			}
		}
	}
}

func main() {
	m := &ShardedMap{}

	m.Put("Alice", 100)
	m.Put("Bob", 200)

	val, found := m.Get("Alice")
	fmt.Println("Alice:", val, "Found:", found)

	m.Delete("Alice")

	val, found = m.Get("Alice")
	fmt.Println("Alice after delete:", val, "Found:", found)
}
