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

	data := make(chan uint64)
	go aximemory.ReadBurstUInt64(
		memReadAddr, memReadData, true, inputData, length, data)
	aximemory.WriteBurstUInt64(
		memWriteAddr, memWriteData, memWriteResp, true, outputData, length, data)
}
