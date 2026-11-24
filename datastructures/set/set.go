// Derived from reqfleet/shibuya; modified for lib.

// Package set provides a concurrency-safe set of comparable values.
package set

import (
	"sync"
)

// Set stores unique values. Create a Set with New; do not copy it after first use.
type Set[T comparable] struct {
	mu    sync.RWMutex
	items map[T]struct{}
}

// New creates an empty Set.
func New[T comparable]() *Set[T] {
	return &Set[T]{
		items: make(map[T]struct{}),
	}
}

// Add inserts item into the set. Adding an existing item has no effect.
func (s *Set[T]) Add(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[item] = struct{}{}
}

// Remove deletes item from the set. Removing an absent item has no effect.
func (s *Set[T]) Remove(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, item)
}

// Contains reports whether item is in the set.
func (s *Set[T]) Contains(item T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.items[item]
	return exists
}

// Size returns the number of items in the set.
func (s *Set[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// Clear removes all items from the set.
func (s *Set[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[T]struct{})
}

// Items returns a snapshot of the set's items in unspecified order.
func (s *Set[T]) Items() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]T, 0, len(s.items))
	for item := range s.items {
		items = append(items, item)
	}
	return items
}
