package main

import (
	//  Import the entire framework for interracting with SDAccel from Go (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	// Use the new AXI protocol package for interracting with memory
	aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
	axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
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
  // YOUR CODE: declare the first operand here
	// YOUR CODE: declare the second operand here
	// YOUR CODE: declare the memory address for the FPGA to store the result

	// Set up channels for interacting with the shared memory
	memReadAddr chan<- axiprotocol.Addr,
	memReadData <-chan axiprotocol.ReadData,

	memWriteAddr chan<- axiprotocol.Addr,
	memWriteData chan<- axiprotocol.WriteData,
	memWriteResp <-chan axiprotocol.WriteResp) {

	// Since we're not reading anything from memory, disable those reads
	go axiprotocol.ReadDisable(memReadAddr, memReadData)

	// Add the two input integers together
	// YOUR CODE: Perform the addition here using the Add function

	// Write the result of the addition to the shared memory address provided by the host
	aximemory.WriteUInt32(
		memWriteAddr, memWriteData, memWriteResp, false, addr, val)
}
