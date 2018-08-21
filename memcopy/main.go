package main

import (
	// Import the entire framework (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"
	"github.com/ReconfigureIO/sdaccel/smi"
)

// Magic identifier for exporting
func Top(
	inputData uintptr,
	outputData uintptr,
	length uint32,

	// Set up channels for interacting with the shared memory
	readReq chan<- smi.Flit64,
	readResp <-chan smi.Flit64,

	writeReq chan<- smi.Flit64,
	writeResp <-chan smi.Flit64) {

	data := make(chan uint64)
	go smi.ReadBurstUInt64(
		readReq, readResp, inputData, smi.DefaultOptions, length, data)
	smi.WriteBurstUInt64(
		writeReq, writeResp, outputData, smi.DefaultOptions, length, data)
}
