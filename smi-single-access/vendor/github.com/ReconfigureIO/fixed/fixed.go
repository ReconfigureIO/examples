// Copyright 2017 Reconfigure.io.
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fixed implements fixed-point integer types for FPGAs
package fixed

type Int26_6 int32

func I26(i int32) Int26_6 {
	return Int26_6(i << 6)
}

func I26F(i int32, f int32) Int26_6 {
	return Int26_6(i<<6 + (f & 0x3f))
}

// The greatest integer value ≤ x.
func (x Int26_6) Floor() int32 {
	return int32(x) >> 6
}

// The nearest integer to x.
func (x Int26_6) Round() int32 {
	return (int32(x) + 0x20) >> 6
}

// The least integer greater than x.
func (x Int26_6) Ceil() int32 {
	return (int32(x) + 0x3f) >> 6
}

// An alias for the builtin addition operation. It is recommended
// that you use the primitive + to avoid the overhead of a function call.
func (x Int26_6) Add(y Int26_6) Int26_6 {
	return x + y
}

// The product of x * y.
// Please note there is no overflow detection at this point.
func (x Int26_6) Mul(y Int26_6) Int26_6 {
	return Int26_6((int64(x)*int64(y) + 1<<5) >> 6)
}

type Int52_12 int64

func I52(x int64) Int52_12 {
	return Int52_12(x << 12)
}

func I52F(x int64, f int64) Int52_12 {
	return Int52_12(x<<12 + (f & 0xfff))
}

// The greatest integer value ≤ x.
func (x Int52_12) Floor() int64 {
	return int64(x) >> 12
}

// The nearest integer to x.
func (x Int52_12) Round() int64 {
	return (int64(x) + 0x800) >> 12
}

// The least integer greater than x.
func (x Int52_12) Ceil() int64 {
	return (int64(x) + 0xfff) >> 12
}

type pair struct {
	low  uint64
	high uint64
}

// muli64 multiplies two int64 values, returning the 128-bit signed integer
// result as two uint64 values.
//
// This implementation is similar to $GOROOT/src/runtime/softfloat64.go's mullu
// function, which is in turn adapted from Hacker's Delight.
func muli64(u int64, v int64) pair {
	const s uint64 = 32
	const mask uint64 = 1<<32 - 1

	u1 := uint64(u >> s)
	u0 := uint64(u & int64(mask))
	v1 := uint64(v >> s)
	v0 := uint64(v & int64(mask))

	w0 := u0 * v0
	t := u1*v0 + w0>>s
	w1 := t & mask
	w2 := uint64(int64(t) >> s)
	w1 += u0 * v1

	return pair{
		low:  uint64(u) * uint64(v),
		high: u1*v1 + w2 + uint64(int64(w1)>>s),
	}
}

// Mul returns x*y in 52.12 fixed-point arithmetic.
func (x Int52_12) Mul(y Int52_12) Int52_12 {
	var M uint64 = 52
	var N uint64 = 12
	result := muli64(int64(x), int64(y))
	lo := result.low
	hi := result.high
	ret := Int52_12(hi<<M | lo>>N)
	ret += Int52_12((lo >> (N - 1)) & 1) // Round to nearest, instead of rounding down.
	return ret
}
