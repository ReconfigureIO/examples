package main

import (
	"fmt"
	"math"

	"github.com/ReconfigureIO/fixed/host"
)

const (
	c = 256
	r = 3.442619855899
)

func f(x float64) float64 {
	return math.Exp(-x * x / 2)
}

func fInv(y float64) float64 {
	return math.Sqrt(-2 * math.Log(y))
}

func printVar(varName string, vals []float64) {
	fmt.Printf("%s = {", varName)
	for i, v := range vals {
		if i != len(vals)-1 {
			fmt.Printf("%d, ", int32(host.I26Float64(v)))
		} else {
			fmt.Printf("%d}\n", int32(host.I26Float64(v)))
		}
	}

}

func main() {
	a := float64(0)
	b := float64(10)
	r := float64(5)
	aa := a
	bb := b
	xs := make([]float64, c, c)
	for aa < r && r < bb {
		aa = a
		bb = b
		r = .5 * (a + b)

		v := r*f(r) + math.Exp(-0.5*r*r)/r
		xs[c-1] = r
		for i := c - 1; i != 0; i-- {
			xs[i-1] = fInv(v/xs[i] + f(xs[i]))
		}
		q := xs[1] * (1 - f(xs[1])) / v
		if q > 1 {
			b = r
		} else {
			a = r
		}
	}
	xs[0] = 0

	ys := make([]float64, c, c)
	for i, x := range xs {
		ys[i] = f(x)
	}
	fs := make([]float64, c, c)
	for i := range ys {
		if i > 0 {
			t := ys[i-1] - ys[i]
			if t < 0 {
				t = -t
			}
			fs[i] = t * 16
		} else {
			fs[i] = 0
		}
	}

	bs := make([]float64, c, c)
	for i := 1; i < c; i++ {
		dy := ys[i] - ys[i-1]
		bs[i] = ((2 / dy) * (dy * fInv((ys[i]+ys[i-1])/2))) - xs[i-1]
	}
	ms := make([]float64, c, c)
	for i := 1; i < c-1; i++ {
		ms[i] = (ys[i] - ys[i+1]) / (xs[i] - bs[i+1])
	}

	conv := func(x float64) uint8 {
		return uint8(int32(host.I26Float64(x)))
	}

	fmt.Printf("%s = {", "params")
	for i, x := range xs {
		xNext := uint8(255)
		if i != len(xs)-1 {
			t := xs[i+1]
			xNext = conv(t)
		}
		m := float64(0)
		if i != 0 {
			m = ms[i-1]
		}

		out := (uint32(conv(x)) << 24)
		out |= uint32(conv(fs[i])) << 16
		out |= uint32(conv(m)) << 8
		out |= uint32(xNext)
		fmt.Printf("%d, ", out)
	}
	fmt.Printf("}\n")
	fmt.Printf("r = %d\n", host.I26Float64(r))
	fmt.Printf("rInv = %d\n", host.I26Float64(1.0/r))

	log2 := math.Log(2)

	const logs = 32
	lns := make([]float64, logs, logs)
	start := float64(0.5)
	for i := 0; i < logs; i++ {
		lns[i] = log2 + math.Log(start)
		start += 0.5 / 32
	}
	printVar("lns", lns)

}
