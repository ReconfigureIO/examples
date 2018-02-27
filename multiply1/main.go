package main

import (
	// Import the entire framework for interracting with SDAccel from Go (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	/// Use the new AXI protocol package for interracting with memory
	aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
	axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
)

func Top(
	// Pass two operands to the FPGA, the integer to be multiplied and a pointer to the
	// space in shared memory where it should store the result.

	a uint32,
	addr uintptr,

	// Set up channels for interacting with the shared memory
	memReadAddr chan<- axiprotocol.Addr,
	memReadData <-chan axiprotocol.ReadData,

	memWriteAddr chan<- axiprotocol.Addr,
	memWriteData chan<- axiprotocol.WriteData,
	memWriteResp <-chan axiprotocol.WriteResp) {

	// Since we're not reading anything from memory, disable those reads
	go axiprotocol.ReadDisable(memReadAddr, memReadData)

	// Multiply incoming data by 2
	val := a * 2

	// Write the result to the location in shared memory as requested by the host
	aximemory.WriteUInt32(
		memWriteAddr, memWriteData, memWriteResp, true, addr, uint32(val))
}
