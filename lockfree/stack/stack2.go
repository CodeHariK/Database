package main

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	val unsafe.Pointer
	nxt unsafe.Pointer
}

func (n *node) value() unsafe.Pointer {
	return atomic.LoadPointer(&n.val)
}

func (n *node) next() *node {
	return (*node)(atomic.LoadPointer(&n.nxt))
}

func (n *node) casNext(expected, target unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&n.nxt, expected, target)
}

func casAddr(addr *unsafe.Pointer, expected, target unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(addr, expected, target)
}

type stack struct {
	count uint64
	head  *node
}

// NewStack creates a new stack
func NewStack() *stack {
	var empty interface{}
	return &stack{
		head: &node{val: unsafe.Pointer(&empty)},
	}
}

func (s *stack) Len() int {
	return int(atomic.LoadUint64(&s.count))
}

func (s *stack) Push(v interface{}) {
	n := node{
		val: unsafe.Pointer(&v),
	}
	headAddr := (*unsafe.Pointer)(unsafe.Pointer(&s.head))
	for {
		head := atomic.LoadPointer(headAddr)
		n.nxt = head
		if casAddr(headAddr, head, unsafe.Pointer(&n)) {
			atomic.AddUint64(&s.count, 1)
			return
		}
	}
}

func (s *stack) Pop() interface{} {
	headAddr := (*unsafe.Pointer)(unsafe.Pointer(&s.head))
	for {
		head := (*node)(atomic.LoadPointer(headAddr))
		n := head.next()
		if n == nil {
			return nil
		}
		if casAddr(headAddr, unsafe.Pointer(head), unsafe.Pointer(n)) {
			atomic.AddUint64(&s.count, ^uint64(0))
			return *(*interface{})(head.value())
		}
	}
}

func (s *stack) Peek() interface{} {
	return *(*interface{})(s.head.value())
}
