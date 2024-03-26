package list

import "sync/atomic"

// Stack is a lockfree implementation of stack (LIFO).
type Stack[T any] struct {
	head  atomic.Pointer[node[T]]
	count atomic.Int64
}

// NewStack returns a Stack (LIFO).
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

// Push an element to the head.
func (s *Stack[T]) Push(v T) {
	node := &node[T]{value: v}
	for {
		head := s.head.Load()
		node.next.Store(head)

		if s.head.CompareAndSwap(head, node) {
			s.count.Add(1)
			return
		}
	}
}

// Pop an element from the head. If stack is empty, ok returns false.
func (s *Stack[T]) Pop() (v T, ok bool) {
	for {
		head := s.head.Load()
		if head == nil {
			return
		}

		if s.head.CompareAndSwap(head, head.next.Load()) {
			s.count.Add(-1)
			return head.value, true
		}
	}
}

// Peek the head element in the stack. If stack is empty, ok returns false.
func (s *Stack[T]) Peek() (v T, ok bool) {
	head := s.head.Load()
	if head == nil {
		return
	}

	return head.value, true
}

// Len returns number of elements in the stack
func (s *Stack[T]) Len() int {
	return int(s.count.Load())
}
