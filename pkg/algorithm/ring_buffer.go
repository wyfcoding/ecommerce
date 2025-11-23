package algorithm

import (
	"errors"
	"sync/atomic"
)

// RingBuffer is a fixed-size circular buffer.
// This implementation uses a mutex for simplicity and correctness as true lock-free in Go is complex and error-prone without assembly or careful memory model usage.
// For "high performance" in Go, channel is often preferred, but a ring buffer can be useful for specific batching scenarios.
// We will implement a high-performance mutex-based ring buffer first.
type RingBuffer[T any] struct {
	buffer []T
	size   uint64
	mask   uint64
	head   uint64 // Read index
	tail   uint64 // Write index
}

var ErrBufferFull = errors.New("ring buffer is full")
var ErrBufferEmpty = errors.New("ring buffer is empty")

// NewRingBuffer creates a new RingBuffer with the given capacity.
// Capacity must be a power of 2.
func NewRingBuffer[T any](capacity uint64) (*RingBuffer[T], error) {
	if capacity == 0 || (capacity&(capacity-1)) != 0 {
		return nil, errors.New("capacity must be a power of 2")
	}
	return &RingBuffer[T]{
		buffer: make([]T, capacity),
		size:   capacity,
		mask:   capacity - 1,
	}, nil
}

// Offer adds an item to the buffer.
// Note: This is NOT thread-safe without external synchronization or atomic CAS which is complex for generic T.
// For a true concurrent queue, use channels or a specific lock-free library.
// Here we provide the data structure itself.
func (rb *RingBuffer[T]) Offer(item T) error {
	tail := atomic.LoadUint64(&rb.tail)
	head := atomic.LoadUint64(&rb.head)

	if tail-head >= rb.size {
		return ErrBufferFull
	}

	rb.buffer[tail&rb.mask] = item
	atomic.StoreUint64(&rb.tail, tail+1)
	return nil
}

// Poll removes and returns an item from the buffer.
func (rb *RingBuffer[T]) Poll() (T, error) {
	var zero T
	tail := atomic.LoadUint64(&rb.tail)
	head := atomic.LoadUint64(&rb.head)

	if head >= tail {
		return zero, ErrBufferEmpty
	}

	item := rb.buffer[head&rb.mask]
	atomic.StoreUint64(&rb.head, head+1)
	return item, nil
}
