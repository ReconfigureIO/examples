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

	readChan := make(chan uintptr)
	readRespChan := make(chan uint32)

	incrChan := make(chan uintptr)
	incrResp := make(chan uint32)

	go func() {
		for {
			var addr uintptr

			// This block provides single access to memory, with two
			// separate operations: reading or incrementing a pointer
			// in memory.

			// All access is through channels allowing operations to
			// be enqueued, only blocking when the response is needed
			select {
			case addr = <-readChan:
				readRespChan <- memory.Read(addr, memReadAddr, memReadData)
			case addr = <-incrChan:
				current := memory.Read(addr, memReadAddr, memReadData)
				memory.Write(addr, current+1, memWriteAddr, memWriteData, memResp)
				incrResp <- current + 1
			}
		}
	}()

	go func() {
		// length is the number of addresses we are supposed to read
		// so this block enqueues each address to read.
		for i := length; i > 0; i-- {
			readChan <- inputData
			inputData += 4
		}
	}()

	go func() {
		for i := length; i > 0; i-- {
			// get the read response that was previously enqueued.
			sample := <-readRespChan
			// If we think of external memory we are writing to as a [512]uint32, this would be the index we access
			index := uint16(sample) >> (16 - 9)
			pointerDiff := index << 2
			// And this is that index as a pointer to external memory
			outputPointer := outputData + uintptr(pointerDiff)
			// enqueue an increment operation on that pointer
			incrChan <- outputPointer
		}
	}()

	// Wait for each response for increment operations
	for i := length; i > 0; i-- {
		<-incrResp
	}

	// Once that's done, we can exit.
}
