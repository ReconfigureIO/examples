package main

import (
	"testing"
	"testing/quick"
)

func TestAdd(t *testing.T) {
	// Check that our Add function adds numbers correctly
	f := func(x uint32, y uint32) bool {
		result := Add(x, y)
		return (result == (x + y))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
