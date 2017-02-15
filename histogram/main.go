package main

import (
	// import the entire framework (including bundled verilog)
	_ "sdaccel"
	"sdaccel/memory"
)

// magic identifier for exporting
func Top(
	inputData uintptr,
	outputData uintptr,
	length uint32,

	memReadAddr chan<- memory.Addr,
	memReadData <-chan memory.ReadData,

	memWriteAddr chan<- memory.Addr,
	memWriteData chan<- memory.WriteData,
	memResp <-chan memory.Response) {

	// The host needs to provide the length we should read
	for ; length > 0; length-- {
		// First we'll read each sample
		sample := memory.Read(inputData, memReadAddr, memReadData)
		// If we think of external memory we are writing to as a [512]uint32, this would be the index we access
		index := uint16(sample) >> (16 - 9)
		pointerDiff := index << 2
		// And this is that index as a pointer to external memory
		outputPointer := outputData + uintptr(pointerDiff)

		current := memory.Read(outputPointer, memReadAddr, memReadData)

		memory.Write(outputPointer, current+1, memWriteAddr, memWriteData, memResp)

		inputData += 4
	}
}
