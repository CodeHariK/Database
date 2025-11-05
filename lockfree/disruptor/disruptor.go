package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

const BufferSize = 1024 // Ring buffer size (power of 2)

type Event struct {
	Value int
}

type RingBuffer struct {
	buffer     [BufferSize]Event
	writeIndex int64
	readIndex  int64
}

func NewRingBuffer() *RingBuffer {
	return &RingBuffer{}
}

// Producer: Writes data to the ring buffer
func (rb *RingBuffer) Publish(value int) {
	for {
		writePos := atomic.LoadInt64(&rb.writeIndex)
		readPos := atomic.LoadInt64(&rb.readIndex)

		// Ensure we don't overwrite unread data (BufferSize ahead)
		if writePos-readPos < BufferSize {
			rb.buffer[writePos%BufferSize] = Event{Value: value}
			atomic.AddInt64(&rb.writeIndex, 1)
			return
		}
	}
}

// Consumer: Reads data from the ring buffer
func (rb *RingBuffer) Consume() (Event, bool) {
	for {
		readPos := atomic.LoadInt64(&rb.readIndex)
		writePos := atomic.LoadInt64(&rb.writeIndex)

		// Ensure data is available to read
		if readPos < writePos {
			event := rb.buffer[readPos%BufferSize]
			atomic.AddInt64(&rb.readIndex, 1)
			return event, true
		}
		return Event{}, false // No new data
	}
}

func main() {
	ringBuffer := NewRingBuffer()

	// Producer
	go func() {
		for i := 0; i < 10; i++ {
			ringBuffer.Publish(i)
			fmt.Println("Produced:", i)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Consumer
	go func() {
		for {
			event, ok := ringBuffer.Consume()
			if ok {
				fmt.Println("Consumed:", event.Value)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	time.Sleep(1 * time.Second) // Let goroutines run
}
