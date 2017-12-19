package main

import (
	// Import the entire framework for interracting with SDAccel from Go (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

  // Use the new AXI protocol package for interracting with memory
	axiarbitrate "github.com/ReconfigureIO/sdaccel/axi/arbitrate"
	aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
	axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
)

func Top(
	// Three operands from the host. Pointers to the input data and the space for the result in shared
	// memory and the length of the input data so the FPGA knows what to expect.
	inputData uintptr,
	outputData uintptr,
	length uint32,

	// Set up channels for interacting with the shared memory
	memReadAddr chan<- axiprotocol.Addr,
	memReadData <-chan axiprotocol.ReadData,

	memWriteAddr chan<- axiprotocol.Addr,
	memWriteData chan<- axiprotocol.WriteData,
	memWriteResp <-chan axiprotocol.WriteResp) {

	readRespChan := make(chan uint32)
	incrRespChan := make(chan uint32)

	// Create a 2-way AXI bus arbiter so that two goroutines can perform
	// concurrent AXI memory reads.
	memReadAddr0 := make(chan axiprotocol.Addr)
	memReadData0 := make(chan axiprotocol.ReadData)
	memReadAddr1 := make(chan axiprotocol.Addr)
	memReadData1 := make(chan axiprotocol.ReadData)
	go axiarbitrate.ReadArbitrateX2(
		memReadAddr, memReadData, memReadAddr0, memReadData0,
		memReadAddr1, memReadData1)

	go func() {
		// Length is the number of addresses we are supposed to read
		// so this block queues reads from each one in turn.
		for i := length; i != 0; i-- {
			readRespChan <- aximemory.ReadUInt32(
				memReadAddr0, memReadData0, true, inputData)
			inputData += 4
		}
	}()

	go func() {
		for i := length; i != 0; i-- {
			// Get the read response that was previously enqueued.
			sample := <-readRespChan
			// If we think of external memory we are writing to as a
			// [512]uint32, this would be the index we access.
			index := uint16(sample) >> (16 - 9)
			// And this is that index as a pointer to external memory.
			outputPointer := outputData + uintptr(index<<2)
			// Perform an increment operation on that location.
			current := aximemory.ReadUInt32(
				memReadAddr1, memReadData1, true, outputPointer)
			current += 1
			aximemory.WriteUInt32(
				memWriteAddr, memWriteData, memWriteResp, true,
				outputPointer, current)
			incrRespChan <- current
		}
	}()

	// Wait for each response for increment operations.
	for i := length; i != 0; i-- {
		<-incrRespChan
	}

	// Once that's done, we can exit.
}
