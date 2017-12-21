package main

import (
	// Import the entire framework for interracting with SDAccel from Go (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	// Use the new AXI protocol package for interracting with memory
	aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
	axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
)

func Top(
	// Pass a pointer to shared memory to tell the FPGA where to find the data to be operated on,
	// and a pointer to the space in shared memory where the result should be stored. Also tell the FPGA
	// the length that the data will be.

	inputData uintptr,
	outputData uintptr,
	length uint32,

	// Set up channels for interacting with the shared memory
	memReadAddr chan<- axiprotocol.Addr,
	memReadData <-chan axiprotocol.ReadData,

	memWriteAddr chan<- axiprotocol.Addr,
	memWriteData chan<- axiprotocol.WriteData,
	memWriteResp <-chan axiprotocol.WriteResp) {

	// Read all the input data into a channel
	inputChan := make(chan uint32)
	go aximemory.ReadBurstUInt32(
		memReadAddr, memReadData, true, inputData, length, inputChan)

	// Create a channel for the result of the calculation
	transformedChan := make(chan uint32)
	// multiply each element of the input channel by 2 and send to the channel we just made to hold the result
	go func() {
		// no need to stop here, which will save us some clocks checking
		for {
			transformedChan <- (<-inputChan) * 2
		}
	}()

	// Write transformed results back to memory
	aximemory.WriteBurstUInt32(
		memWriteAddr, memWriteData, memWriteResp, true, outputData, length, transformedChan)
}
