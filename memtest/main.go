package main

import (
	// import the entire framework (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
	axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
)

// magic identifier for exporting
func Top(
	outputData uintptr,

	memReadAddr chan<- axiprotocol.Addr,
	memReadData <-chan axiprotocol.ReadData,

	memWriteAddr chan<- axiprotocol.Addr,
	memWriteData chan<- axiprotocol.WriteData,
	memResp <-chan axiprotocol.WriteResp) {

	go aximemory.DisableReads(memReadAddr, memReadData)

	for out := uint64(0); out < 4294967295; out += 4 {
		aximemory.Write(outputData, 1, memWriteAddr, memWriteData, memResp)
	}

}
