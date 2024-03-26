package list

import (
	"sync/atomic"
)

// Queue is a lockfree implementation of queue (FIFO).
type Queue[T any] struct {
	head, tail atomic.Pointer[node[T]]
	count      atomic.Int64
}

// NewQueue returns a Queue (FIFO).
func NewQueue[T any]() *Queue[T] {
	queue := &Queue[T]{}
	placeholder := &node[T]{}
	queue.head.Store(placeholder)
	queue.tail.Store(placeholder)
	return queue
}

// Enqueue an element at the end of the queue.
func (q *Queue[T]) Enqueue(v T) {
	node := &node[T]{value: v}
	for {
		tail := q.tail.Load()

		if q.tail.CompareAndSwap(tail, node) {
			tail.next.Store(node)
			q.count.Add(1)
			return
		}
	}
}

// Dequeue removes and returns the head element of the queue. If queue is empty, ok returns false.
func (q *Queue[T]) Dequeue() (v T, ok bool) {
	for {
		head := q.head.Load()
		next := head.next.Load()
		if next == nil {
			return
		}

		if q.head.CompareAndSwap(head, next) {
			q.count.Add(-1)
			return next.value, true
		}
	}
}

// Peek the head element in the queue without removing. If queue is empty, ok returns false.
func (q *Queue[T]) Peek() (v T, ok bool) {
	head := q.head.Load()
	next := head.next.Load()
	if next == nil {
		return
	}

	return next.value, true
}

// Len returns number of elements in the queue.
func (q *Queue[T]) Len() int {
	return int(q.count.Load())
}
