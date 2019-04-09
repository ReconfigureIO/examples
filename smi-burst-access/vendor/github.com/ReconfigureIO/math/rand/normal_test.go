package rand

import (
	"math"
	"testing"

	"github.com/ReconfigureIO/fixed"
)

func TestNormals(t *testing.T) {
	r := New(42)
	out := make(chan fixed.Int26_6)

	const iterations = 1024 * 1024

	go r.Normals(out)

	var sums, squares float64
	for i := 0; i < iterations; i++ {
		o := float64(<-out) / float64(1<<6)
		sums += o
		squares += (o * o)

	}
	mean := sums / iterations
	stddev := math.Sqrt((iterations*squares - sums*sums) / (iterations * (iterations - 1)))
	if math.Abs(mean) > 0.001 || math.Abs(1-stddev) > 0.2 {
		t.Errorf("Expected a mean of ~0 & stddev of ~1, got %f & %f", mean, stddev)
	}
}
