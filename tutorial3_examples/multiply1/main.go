package main

import (
	// Import the entire framework (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	// Use the new AXI protocol package
	aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
	axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
)

func Top(
	// Specify inputs and outputs to the kernel. Tell the kernel where to find data in shared memory, what data type
	// to expect or pass single integers directly to the kernel by sending them to the FPGA's control register

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

	// Do whatever needs doing with the data from the host

	// Multiply incoming data by 2
	val := a * 2

	// Write the result to the location in shared memory as requested by the host
	aximemory.WriteUInt32(
		memWriteAddr, memWriteData, memWriteResp, true, addr, uint32(val))
}
