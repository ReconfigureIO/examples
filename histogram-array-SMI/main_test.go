package main

import (
	"testing"
	"testing/quick"
)

func TestCalculateIndexDoesNotOutOfBounds(t *testing.T) {
	// Check that we never generate an index out of bounds
	f := func(x uint32) bool {
		index := CalculateIndex(x)
		return index < 512
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
