package main

import (
	// Import the entire framework (including bundled verilog)
	_ "sdaccel"
	// Use the new AXI protocol package
	aximemory "axi/memory"
	axiprotocol "axi/protocol"
)

// Magic identifier for exporting
func Top(
	inputData uintptr,
	outputData uintptr,
	length uint32,

	memReadAddr chan<- axiprotocol.Addr,
	memReadData <-chan axiprotocol.ReadData,

	memWriteAddr chan<- axiprotocol.Addr,
	memWriteData chan<- axiprotocol.WriteData,
	memWriteResp <-chan axiprotocol.WriteResp) {

	// The host needs to provide the length we should read
	for ; length > 0; length-- {
		// First we'll read each sample
		sample := aximemory.ReadUInt32(
			memReadAddr, memReadData, true, inputData)

		// If we think of external memory we are writing to as a
		// [512]uint32, this would be the index we access, with an
		// extra 2 bits to make it a pointer
		outputPointer := outputData + uintptr((uint16(sample)>>(16-9))<<2)

		current := aximemory.ReadUInt32(
			memReadAddr, memReadData, true, outputPointer)

		aximemory.WriteUInt32(
			memWriteAddr, memWriteData, memWriteResp, true,
			outputPointer, current+1)

		inputData += 4
	}
}
