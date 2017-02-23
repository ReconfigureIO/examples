package main

import (
	// import the entire framework (including bundled verilog)
	_ "sdaccel"

	"sdaccel/memory"
)

// The Top function will be presented as a kernel
func Top(
	// The first set of arguments to this function can be any number
	// of Go primitive types and can be provided via `SetArg` on the host.

	// For this example, we have 3 arguments: two operands to add
	// together and an address to memory (the uint32s) on the FPGA to
	// store the output (the uintptr)
	a uint32,
	b uint32,
	addr uintptr,

	// The second set of arguments will be the ports for interacting with memory
	memReadAddr chan<- memory.Addr,
	memReadData <-chan memory.ReadData,

	memWriteAddr chan<- memory.Addr,
	memWriteData chan<- memory.WriteData,
	memResp <-chan memory.Response) {

	// Since we're not reading anything from memory, disable those reads
	go memory.DisableReads(memReadAddr, memReadData)

	// Calculate the value
	val := a + b

	// Write it back to the pointer the host requests
	memory.Write(addr, val, memWriteAddr, memWriteData, memResp)
}
