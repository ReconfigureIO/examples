package main

import (
	// Import the entire framework for interracting with SDAccel from Go (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	// Use the SMI protocol package
	"github.com/ReconfigureIO/sdaccel/smi"
)

func Top(
	// For this example, we have 3 arguments: Pointers to the input data, the
	// space for the result and the length of the input data so the FPGA knows
	// what to expect.
	inputData uintptr,
	outputData uintptr,
	length uint32,

	// Set up SMI ports for interacting with the shared memory,
	// two read ports and one write port

	readAReq chan<- smi.Flit64,
	readAResp <-chan smi.Flit64,

	readBReq chan<- smi.Flit64,
	readBResp <-chan smi.Flit64,

	writeReq chan<- smi.Flit64,
	writeResp <-chan smi.Flit64) {

	readRespChan := make(chan uint32)
	incrRespChan := make(chan uint32)

	go func() {
		// Length is the number of addresses we are supposed to read
		// so this block queues reads from each one in turn.

		for i := length; i != 0; i-- {
			readRespChan <- smi.ReadUInt32(
				readAReq, readAResp, inputData, smi.DefaultOptions)
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
			current := smi.ReadUInt32(
				readBReq, readBResp, outputPointer, smi.DefaultOptions)
			current += 1
			smi.WriteUInt32(
				writeReq, writeResp, outputPointer, smi.DefaultOptions, current)
			incrRespChan <- current
		}
	}()

	// Wait for each response for increment operations.
	for i := length; i != 0; i-- {
		<-incrRespChan
	}

	// Once that's done, we can exit.
}
