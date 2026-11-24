package stack

import "testing"

func TestStack_Pop_ReleasesReferences(t *testing.T) {
	s := New[*int]()
	values := []int{1, 2}
	for i := range values {
		s.Push(&values[i])
	}
	backing := s.items

	for i := len(values) - 1; i >= 0; i-- {
		item, ok := s.Pop()
		if !ok || item != &values[i] {
			t.Fatalf("Pop() = (%v, %v), want (%v, true)", item, ok, &values[i])
		}
		// Inspect the backing array rather than relying on GC timing.
		if backing[i] != nil {
			t.Errorf("popped slot %d retains a reference", i)
		}
	}
}
