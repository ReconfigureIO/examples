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

	for ; length > 0; length-- {
		sample := memory.Read(inputData, memReadAddr, memReadData)
		memory.Write(outputData, sample, memWriteAddr, memWriteData, memResp)
		inputData += 4
		outputData += 4

	}

}
