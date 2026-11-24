// Derived from reqfleet/shibuya; modified for lib.
package stack_test

import (
	"testing"

	"github.com/reqfleet/lib/datastructures/stack"
)

func TestStack_Pop_EmptyStack(t *testing.T) {
	stack := stack.New[any]()
	item, ok := stack.Pop()
	if ok {
		t.Errorf("expected ok to be false for empty stack, got true")
	}
	if item != nil {
		t.Errorf("expected item to be nil for empty stack, got %v", item)
	}
}

func TestStack_Pop_NonEmptyStack(t *testing.T) {
	stack := stack.New[any]()
	stack.Push(42)
	stack.Push("hello")

	item, ok := stack.Pop()
	if !ok {
		t.Errorf("expected ok to be true, got false")
	}
	if item != "hello" {
		t.Errorf("expected 'hello', got %v", item)
	}

	item, ok = stack.Pop()
	if !ok {
		t.Errorf("expected ok to be true, got false")
	}
	if item != 42 {
		t.Errorf("expected 42, got %v", item)
	}

	item, ok = stack.Pop()
	if ok {
		t.Errorf("expected ok to be false after popping all items, got true")
	}
	if item != nil {
		t.Errorf("expected item to be nil after popping all items, got %v", item)
	}
}

func TestStack_Pop_Concurrent(t *testing.T) {
	stack := stack.New[any]()
	for i := range 100 {
		stack.Push(i)
	}

	done := make(chan struct{})
	go func() {
		for range 100 {
			stack.Pop()
		}
		close(done)
	}()
	<-done
	item, ok := stack.Pop()
	if ok {
		t.Errorf("expected ok to be false after concurrent pops, got true")
	}
	if item != nil {
		t.Errorf("expected item to be nil after concurrent pops, got %v", item)
	}
}
