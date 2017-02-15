package main

import (
	// import the entire framework (including bundled verilog)
	_ "sdaccel"

	"sdaccel/memory"
)

// magic identifier for exporting
func Top(
	a uint32,
	b uint32,
	addr uintptr,

	memReadAddr chan<- memory.Addr,
	memReadData <-chan memory.ReadData,

	memWriteAddr chan<- memory.Addr,
	memWriteData chan<- memory.WriteData,
	memResp <-chan memory.Response) {

	// Disable memory reads
	go memory.DisableReads(memReadAddr, memReadData)

	val := a + b

	memory.Write(addr, val, memWriteAddr, memWriteData, memResp)
}
