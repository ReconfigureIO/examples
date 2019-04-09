package host

import (
	"testing"
)

func TestI26F(t *testing.T) {
	f := I26Float64(9.664649023)
	if f.Floor() != 9 {
		t.Errorf("Expected %d, got %d", 9, f.Floor())
	}

	f = I26Float64(-9.664649023)
	if f.Floor() != -10 {
		t.Errorf("Expected %d, got %d", -10, f.Floor())
	}
}

func TestI52F(t *testing.T) {
	f := I52Float64(9.664649023)
	if f.Floor() != 9 {
		t.Errorf("Expected %d, got %d", 9, f.Floor())
	}

	f = I52Float64(-9.664649023)
	if f.Floor() != -10 {
		t.Errorf("Expected %d, got %d", -10, f.Floor())
	}
}
