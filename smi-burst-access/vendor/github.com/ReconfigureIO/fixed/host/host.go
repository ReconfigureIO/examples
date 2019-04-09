package host

import (
	"github.com/ReconfigureIO/fixed"
)

func I26Float64(f float64) fixed.Int26_6 {
	return fixed.Int26_6(f * (1 << 6))
}

func I52Float64(f float64) fixed.Int52_12 {
	return fixed.Int52_12(f * (1 << 12))
}
