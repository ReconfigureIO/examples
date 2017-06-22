package main

import (
	// import the entire framework (including bundled verilog)
	_ "sdaccel"
	// Use the new AXI protocol package
	aximemory "axi/memory"
	axiprotocol "axi/protocol"
)

// magic identifier for exporting
func Top(
	inputData uintptr,
	outputData uintptr,
	length uint32,

	memReadAddr chan<- axiprotocol.Addr,
	memReadData <-chan axiprotocol.ReadData,

	memWriteAddr chan<- axiprotocol.Addr,
	memWriteData chan<- axiprotocol.WriteData,
	memWriteResp <-chan axiprotocol.WriteResp) {

	var histogram [512]uint32

	// Read all of the input data into a channel
	inputChan := make(chan uint32)
	go aximemory.ReadBurstUInt32(
		memReadAddr, memReadData, true, inputData, length, inputChan)

	// The host needs to provide the length we should read
	for ; length > 0; length-- {
		// First we'll pull of each sample from the channel
		sample := <-inputChan
		// calculate the bin for the histogram
		index := uint16(sample) >> (16 - 9)

		// And increment the value in that bin
		histogram[uint(index)] += 1
	}

	data := make(chan uint32)
	go func() {
		for i := 0; i < 512; i++ {
			data <- histogram[i]
		}
	}()

	aximemory.WriteBurstUInt32(
		memWriteAddr, memWriteData, memWriteResp, true, outputData, 512, data)
}
