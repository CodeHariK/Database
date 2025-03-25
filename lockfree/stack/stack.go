package main

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type Node struct {
	value int
	next  *Node
}

type LockFreeStack struct {
	top unsafe.Pointer
}

func (s *LockFreeStack) Push(value int) {
	newNode := &Node{value: value}

	for {
		oldTop := (*Node)(atomic.LoadPointer(&s.top))
		newNode.next = oldTop

		if atomic.CompareAndSwapPointer(&s.top, unsafe.Pointer(oldTop), unsafe.Pointer(newNode)) {
			break
		}
	}
}

func (s *LockFreeStack) Pop() (int, bool) {
	for {
		oldTop := (*Node)(atomic.LoadPointer(&s.top))
		if oldTop == nil {
			return 0, false // Stack is empty
		}

		newTop := oldTop.next
		if atomic.CompareAndSwapPointer(&s.top, unsafe.Pointer(oldTop), unsafe.Pointer(newTop)) {
			return oldTop.value, true
		}
	}
}

func main() {
	stack := LockFreeStack{}
	stack.Push(10)
	stack.Push(20)

	val, _ := stack.Pop()
	fmt.Println("Popped:", val)

	val, _ = stack.Pop()
	fmt.Println("Popped:", val)
}
