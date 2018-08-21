package main

import (
	//  Import the entire framework for interracting with SDAccel from Go (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	// Use the SMI protocol package
	"github.com/ReconfigureIO/sdaccel/smi"
)

// function to add two uint32s
func Add(a uint32, b uint32) uint32 {
	return a + b
}

func Top(
	// The first set of arguments to this function can be any number
	// of Go primitive types and can be provided via `SetArg` on the host.

	// For this example, we have 3 arguments: two operands to add
	// together and an address in shared memory where the FPGA will
	// store the output.
	a uint32,
	b uint32,
	addr uintptr,

	// Set up channel to write result to shared memory
	writeReq chan<- smi.Flit64,
	writeResp <-chan smi.Flit64) {

	// Add the two input integers together
	val := Add(a, b)

	// Write the result of the addition to the shared memory address provided by the host
	smi.WriteUInt32(
		writeReq, writeResp, addr, 32, val)
}
