package main

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type Node struct {
	value int
	next  unsafe.Pointer
}

type LockFreeQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

func NewQueue() *LockFreeQueue {
	dummy := &Node{}
	return &LockFreeQueue{
		head: unsafe.Pointer(dummy),
		tail: unsafe.Pointer(dummy),
	}
}

func (q *LockFreeQueue) Enqueue(value int) {
	newNode := &Node{value: value}

	for {
		tail := (*Node)(atomic.LoadPointer(&q.tail))
		next := (*Node)(atomic.LoadPointer(&tail.next))

		if next != nil {
			atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			continue
		}

		if atomic.CompareAndSwapPointer(&tail.next, nil, unsafe.Pointer(newNode)) {
			atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(newNode))
			return
		}
	}
}

func (q *LockFreeQueue) Dequeue() (int, bool) {
	for {
		head := (*Node)(atomic.LoadPointer(&q.head))
		// tail := (*Node)(atomic.LoadPointer(&q.tail))
		next := (*Node)(atomic.LoadPointer(&head.next))

		if next == nil {
			return 0, false // Queue is empty
		}

		if atomic.CompareAndSwapPointer(&q.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
			return next.value, true
		}
	}
}

func main() {
	queue := NewQueue()
	queue.Enqueue(1)
	queue.Enqueue(2)

	val, _ := queue.Dequeue()
	fmt.Println("Dequeued:", val)

	val, _ = queue.Dequeue()
	fmt.Println("Dequeued:", val)
}
