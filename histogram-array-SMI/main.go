package main

import (
	// import the entire framework (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	// Use the new SMI protocol package
	"github.com/ReconfigureIO/sdaccel/smi"
)

// magic identifier for exporting
func Top(
	inputData uintptr,
	outputData uintptr,
	length uint32,

	readReq chan<- smi.Flit64,
	readResp <-chan smi.Flit64,

	writeReq chan<- smi.Flit64,
	writeResp <-chan smi.Flit64) {

	histogram := [512]uint32{}

	// Read all of the input data into a channel
	inputChan := make(chan uint32)
	go smi.ReadBurstUInt32(readReq, readResp, inputData, smi.DefaultOptions, length, inputChan)

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

	smi.WriteBurstUInt32(
		writeReq, writeResp, outputData, smi.DefaultOptions, 512, data)
}
