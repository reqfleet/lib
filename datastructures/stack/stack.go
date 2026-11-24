// Derived from reqfleet/shibuya; modified for lib.

// Package stack provides a concurrency-safe last-in, first-out stack.
package stack

import (
	"sync"
)

// Stack is a thread-safe stack of any values.
// The zero value is ready for use. A Stack must not be copied after first use.
type Stack[T any] struct {
	mu    sync.Mutex
	items []T
}

// New creates a new Stack.
func New[T any]() *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0),
	}
}

// Push adds an item to the top of the stack.
func (s *Stack[T]) Push(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

// Pop removes and returns the item from the top of the stack.
// Returns the zero value of T and false if the stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	idx := len(s.items) - 1
	item := s.items[idx]
	s.items = s.items[:idx]
	return item, true
}
