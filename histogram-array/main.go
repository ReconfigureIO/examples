package main

import (
	// Import the entire framework for interracting with SDAccel from Go (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	// Use the new AXI protocol package for interracting with memory
	aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
	axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
)

// function to calculate the bin for each sample
func CalculateIndex(sample uint32) uint16 {
	return uint16(sample) >> (16 - 9)
}

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

  // Create an array to hold the histogram data as it is sorted
	var histogram [512]uint32

	// Read all of the input data into a channel
	inputChan := make(chan uint32)
	go aximemory.ReadBurstUInt32(
		memReadAddr, memReadData, true, inputData, length, inputChan)

	// A for loop to calculate the histogram data. The host provides the length we should read
	for ; length > 0; length-- {
		// First we'll pull off each sample from the channel
		sample := <-inputChan

		// And increment the value in the correct bin using the calculation function
		histogram[CalculateIndex(sample)] += 1
	}

	// Write the results to a new channel
	data := make(chan uint32)
	go func() {
		for i := 0; i < 512; i++ {
			data <- histogram[i]
		}
	}()

	// Write the results to shared memory
	aximemory.WriteBurstUInt32(
		memWriteAddr, memWriteData, memWriteResp, true, outputData, 512, data)
}
