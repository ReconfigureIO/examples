package main

import (
	// import the entire framework (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"
	"github.com/ReconfigureIO/sdaccel/axi/memory"
)

// magic identifier for exporting
func Top(
	outputData uintptr,

	memReadAddr chan<- memory.Addr,
	memReadData <-chan memory.ReadData,

	memWriteAddr chan<- memory.Addr,
	memWriteData chan<- memory.WriteData,
	memResp <-chan memory.Response) {

	go memory.DisableReads(memReadAddr, memReadData)

	for out := uint64(0); out < 4294967295; out += 4 {
		memory.Write(outputData, 1, memWriteAddr, memWriteData, memResp)
	}

}
