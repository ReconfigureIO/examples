package main

import (
	// Import the entire framework (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
	axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
)

// The Top function will be presented as a kernel
func Top(
	// The first set of arguments to this function can be any number
	// of Go primitive types and can be provided via `SetArg` on the host.

	// For this example, we have 3 arguments: two operands to add
	// together and an address to memory (the uint32s) on the FPGA to
	// store the output (the uintptr)
	// YOUR CODE: declare the first operand here
	// YOUR CODE: declare the second operand here
	// YOUR CODE: declare the memory address for the FPGA to store the result

	// The second set of arguments will be the ports for interacting with memory
	memReadAddr chan<- axiprotocol.Addr,
	memReadData <-chan axiprotocol.ReadData,

	memWriteAddr chan<- axiprotocol.Addr,
	memWriteData chan<- axiprotocol.WriteData,
	memWriteResp <-chan axiprotocol.WriteResp) {

	// Since we're not reading anything from memory, disable those reads
	go axiprotocol.ReadDisable(memReadAddr, memReadData)

	// Calculate the value
	// YOUR CODE: Perform the addition here

	// Write it back to the pointer the host requests
	aximemory.WriteUInt32(
		memWriteAddr, memWriteData, memWriteResp, false, addr, val)
}
