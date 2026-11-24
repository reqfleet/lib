// Derived from reqfleet/shibuya; modified for lib.
package set_test

import (
	"testing"

	"github.com/reqfleet/lib/datastructures/set"
	"github.com/stretchr/testify/assert"
)

func TestSet_Add(t *testing.T) {
	s := set.New[int]()
	s.Add(1)
	assert.True(t, s.Contains(1))
	assert.Equal(t, 1, s.Size())

	s.Add(2)
	assert.True(t, s.Contains(2))
	assert.Equal(t, 2, s.Size())

	// Adding the same item again should not increase the size
	s.Add(1)
	assert.Equal(t, 2, s.Size())
}
