package main

import (
	// Import the entire framework (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	// Use the new SMI protocol package
	"github.com/ReconfigureIO/sdaccel/smi"

	// Use the math package random number source
	"github.com/ReconfigureIO/math/rand"
)

// Structure for holding individual test results.
type resultType struct {
	byteCount  uint32
	errorCount uint32
}

// Function for writing the specified number of counter values to successive
// 8-bit memory locations.
func writeSeqUint8(smiRequest chan<- smi.Flit64, smiResponse <-chan smi.Flit64,
	baseAddr uintptr, length uint32, initVal uint8, incrVal uint8) {

	writeAddr := baseAddr
	writeData := initVal
	for i := length; i != 0; i-- {
		smi.WriteUInt8(smiRequest, smiResponse, writeAddr,
			smi.DefaultOptions, writeData)
		writeAddr += 1
		writeData += incrVal
	}
}

// Function for writing the specified number of counter values to successive
// 16-bit memory locations.
func writeSeqUint16(smiRequest chan<- smi.Flit64, smiResponse <-chan smi.Flit64,
	baseAddr uintptr, length uint32, initVal uint16, incrVal uint16) {

	writeAddr := baseAddr
	writeData := initVal
	for i := length; i != 0; i-- {
		smi.WriteUInt16(smiRequest, smiResponse, writeAddr,
			smi.DefaultOptions, writeData)
		writeAddr += 2
		writeData += incrVal
	}
}

// Function for writing the specified number of counter values to successive
// 32-bit memory locations.
func writeSeqUint32(smiRequest chan<- smi.Flit64, smiResponse <-chan smi.Flit64,
	baseAddr uintptr, length uint32, initVal uint32, incrVal uint32) {

	writeAddr := baseAddr
	writeData := initVal
	for i := length; i != 0; i-- {
		smi.WriteUInt32(smiRequest, smiResponse, writeAddr,
			smi.DefaultOptions, writeData)
		writeAddr += 4
		writeData += incrVal
	}
}

// Function for writing the specified number of counter values to successive
// 64-bit memory locations.
func writeSeqUint64(smiRequest chan<- smi.Flit64, smiResponse <-chan smi.Flit64,
	baseAddr uintptr, length uint32, initVal uint64, incrVal uint64) {

	writeAddr := baseAddr
	writeData := initVal
	for i := length; i != 0; i-- {
		smi.WriteUInt64(smiRequest, smiResponse, writeAddr,
			smi.DefaultOptions, writeData)
		writeAddr += 8
		writeData += incrVal
	}
}

// Function for checking the specified number of counter values in successive
// 8-bit memory locations.
func checkSeqUint8(smiRequest chan<- smi.Flit64, smiResponse <-chan smi.Flit64,
	baseAddr uintptr, length uint32, initVal uint8, incrVal uint8) uint32 {

	readAddr := baseAddr
	checkData := initVal
	errorCount := uint32(0)
	for i := length; i != 0; i-- {
		readData := smi.ReadUInt8(smiRequest, smiResponse, readAddr,
			smi.DefaultOptions)
		if readData != checkData {
			errorCount += 1
		}
		readAddr += 1
		checkData += incrVal
	}
	return errorCount
}

// Function for checking the specified number of counter values in successive
// 16-bit memory locations.
func checkSeqUint16(smiRequest chan<- smi.Flit64, smiResponse <-chan smi.Flit64,
	baseAddr uintptr, length uint32, initVal uint16, incrVal uint16) uint32 {

	readAddr := baseAddr
	checkData := initVal
	errorCount := uint32(0)
	for i := length; i != 0; i-- {
		readData := smi.ReadUInt16(smiRequest, smiResponse, readAddr,
			smi.DefaultOptions)
		if readData != checkData {
			errorCount += 1
		}
		readAddr += 2
		checkData += incrVal
	}
	return errorCount
}

// Function for checking the specified number of counter values in successive
// 32-bit memory locations.
func checkSeqUint32(smiRequest chan<- smi.Flit64, smiResponse <-chan smi.Flit64,
	baseAddr uintptr, length uint32, initVal uint32, incrVal uint32) uint32 {

	readAddr := baseAddr
	checkData := initVal
	errorCount := uint32(0)
	for i := length; i != 0; i-- {
		readData := smi.ReadUInt32(smiRequest, smiResponse, readAddr,
			smi.DefaultOptions)
		if readData != checkData {
			errorCount += 1
		}
		readAddr += 4
		checkData += incrVal
	}
	return errorCount
}

// Function for checking the specified number of counter values in successive
// 64-bit memory locations.
func checkSeqUint64(smiRequest chan<- smi.Flit64, smiResponse <-chan smi.Flit64,
	baseAddr uintptr, length uint32, initVal uint64, incrVal uint64) uint32 {

	readAddr := baseAddr
	checkData := initVal
	errorCount := uint32(0)
	for i := length; i != 0; i-- {
		readData := smi.ReadUInt64(smiRequest, smiResponse, readAddr,
			smi.DefaultOptions)
		if readData != checkData {
			errorCount += 1
		}
		readAddr += 8
		checkData += incrVal
	}
	return errorCount
}

// Run the specified number of 8-bit memory access tests.
func runTestUint8(readUint8Req chan<- smi.Flit64, readUint8Resp <-chan smi.Flit64,
	writeUint8Req chan<- smi.Flit64, writeUint8Resp <-chan smi.Flit64,
	workspacePtr uintptr, workspaceSize uint32, numTransfers uint32,
	resultChan chan<- resultType) {

	result := resultType{0, 0}
	randSource := rand.New(uint32(workspacePtr) | 1)
	randValues := make(chan uint32, 2)
	randSource.Uint32s(randValues)
	for i := numTransfers; i != 0; i-- {
		transferOffset := <-randValues
		transferOffset %= workspaceSize / 2
		transferLength := <-randValues
		transferLength %= (workspaceSize - transferOffset)
		baseAddr := workspacePtr + uintptr(transferOffset)
		initVal := uint8(<-randValues)
		incrVal := uint8(<-randValues)
		writeSeqUint8(writeUint8Req, writeUint8Resp, baseAddr,
			transferLength, initVal, incrVal)
		errorCount := checkSeqUint8(readUint8Req, readUint8Resp,
			baseAddr, transferLength, initVal, incrVal)
		result.byteCount += transferLength
		result.errorCount += errorCount
	}
	resultChan <- result
}

// Run the specified number of 16-bit memory access tests.
func runTestUint16(readUint16Req chan<- smi.Flit64, readUint16Resp <-chan smi.Flit64,
	writeUint16Req chan<- smi.Flit64, writeUint16Resp <-chan smi.Flit64,
	workspacePtr uintptr, workspaceSize uint32, numTransfers uint32,
	resultChan chan<- resultType) {

	result := resultType{0, 0}
	randSource := rand.New(uint32(workspacePtr) | 1)
	randValues := make(chan uint32, 2)
	randSource.Uint32s(randValues)
	for i := numTransfers; i != 0; i-- {
		transferOffset := <-randValues
		transferOffset %= workspaceSize / 2
		transferLength := <-randValues
		transferLength %= (workspaceSize - transferOffset) / 2
		baseAddr := workspacePtr + uintptr(transferOffset)
		initVal := uint16(<-randValues)
		incrVal := uint16(<-randValues)
		writeSeqUint16(writeUint16Req, writeUint16Resp, baseAddr,
			transferLength, initVal, incrVal)
		errorCount := checkSeqUint16(readUint16Req, readUint16Resp,
			baseAddr, transferLength, initVal, incrVal)
		result.byteCount += transferLength * 2
		result.errorCount += errorCount
	}
	resultChan <- result
}

// Run the specified number of 32-bit memory access tests.
func runTestUint32(readUint32Req chan<- smi.Flit64, readUint32Resp <-chan smi.Flit64,
	writeUint32Req chan<- smi.Flit64, writeUint32Resp <-chan smi.Flit64,
	workspacePtr uintptr, workspaceSize uint32, numTransfers uint32,
	resultChan chan<- resultType) {

	result := resultType{0, 0}
	randSource := rand.New(uint32(workspacePtr) | 1)
	randValues := make(chan uint32, 2)
	randSource.Uint32s(randValues)
	for i := numTransfers; i != 0; i-- {
		transferOffset := <-randValues
		transferOffset %= workspaceSize / 2
		transferLength := <-randValues
		transferLength %= (workspaceSize - transferOffset) / 4
		baseAddr := workspacePtr + uintptr(transferOffset)
		initVal := uint32(<-randValues)
		incrVal := uint32(<-randValues)
		writeSeqUint32(writeUint32Req, writeUint32Resp, baseAddr,
			transferLength, initVal, incrVal)
		errorCount := checkSeqUint32(readUint32Req, readUint32Resp,
			baseAddr, transferLength, initVal, incrVal)
		result.byteCount += transferLength * 4
		result.errorCount += errorCount
	}
	resultChan <- result
}

// Run the specified number of 64-bit memory access tests.
func runTestUint64(readUint64Req chan<- smi.Flit64, readUint64Resp <-chan smi.Flit64,
	writeUint64Req chan<- smi.Flit64, writeUint64Resp <-chan smi.Flit64,
	workspacePtr uintptr, workspaceSize uint32, numTransfers uint32,
	resultChan chan<- resultType) {

	result := resultType{0, 0}
	randSource := rand.New(uint32(workspacePtr) | 1)
	randValues := make(chan uint32, 2)
	randSource.Uint32s(randValues)
	for i := numTransfers; i != 0; i-- {
		transferOffset := <-randValues
		transferOffset %= workspaceSize / 2
		transferLength := <-randValues
		transferLength %= (workspaceSize - transferOffset) / 8
		baseAddr := workspacePtr + uintptr(transferOffset)
		initVal := uint64(<-randValues)
		incrVal := uint64(<-randValues)
		writeSeqUint64(writeUint64Req, writeUint64Resp, baseAddr,
			transferLength, initVal, incrVal)
		errorCount := checkSeqUint64(readUint64Req, readUint64Resp,
			baseAddr, transferLength, initVal, incrVal)
		result.byteCount += transferLength * 8
		result.errorCount += errorCount
	}
	resultChan <- result
}

// Top level with multiple SMI interfaces.
func Top(
	// Pointer to memory test workspace area
	workspacePtr uintptr,
	// Size of memory test workspace area
	workspaceSize uint32,
	// Number of write/read sequences
	numTransfers uint32,
	// Pointer to 64-bit byte count result
	byteCountPtr uintptr,
	// Pointer to 64-bit error count result
	errorCountPtr uintptr,

	// SMI read and write channels for 8 bit access tests.
	readUint8Req chan<- smi.Flit64,
	readUint8Resp <-chan smi.Flit64,
	writeUint8Req chan<- smi.Flit64,
	writeUint8Resp <-chan smi.Flit64,

	// SMI read and write channels for 16 bit access tests.
	readUint16Req chan<- smi.Flit64,
	readUint16Resp <-chan smi.Flit64,
	writeUint16Req chan<- smi.Flit64,
	writeUint16Resp <-chan smi.Flit64,

	// SMI read and write channels for 32 bit access tests.
	readUint32Req chan<- smi.Flit64,
	readUint32Resp <-chan smi.Flit64,
	writeUint32Req chan<- smi.Flit64,
	writeUint32Resp <-chan smi.Flit64,

	// SMI read and write channels for 64 bit access tests.
	readUint64Req chan<- smi.Flit64,
	readUint64Resp <-chan smi.Flit64,
	writeUint64Req chan<- smi.Flit64,
	writeUint64Resp <-chan smi.Flit64,

	// SMI write channels for result outputs.
	writeResultReq chan<- smi.Flit64,
	writeResultResp <-chan smi.Flit64,
) {
	byteCount := uint64(0)
	errorCount := uint64(0)

	// Divide workspace area up according to transfer size.
	// Calculate the workspace base pointers on the assumption that the base
	// pointer is aligned to a 64-bit work boundary.
	workspaceSizeUint64 := (workspaceSize / 2) & 0xFFFFFFF8
	workspaceSizeUint32 := (workspaceSize / 4) & 0xFFFFFFFC
	workspaceSizeUint16 := (workspaceSize / 8) & 0xFFFFFFFE
	workspaceSizeUint8 := workspaceSize -
		(workspaceSizeUint64 + workspaceSizeUint32 + workspaceSizeUint16)

	workspacePtrUint64 := workspacePtr
	workspacePtrUint32 := workspacePtrUint64 + uintptr(workspaceSizeUint64)
	workspacePtrUint16 := workspacePtrUint32 + uintptr(workspaceSizeUint32)
	workspacePtrUint8 := workspacePtrUint16 + uintptr(workspaceSizeUint16)

	// Create channels for test result return values.
	resultChanUint8 := make(chan resultType, 1)
	resultChanUint16 := make(chan resultType, 1)
	resultChanUint32 := make(chan resultType, 1)
	resultChanUint64 := make(chan resultType, 1)

	// Run the tests in parallel.
	go runTestUint8(readUint8Req, readUint8Resp, writeUint8Req, writeUint8Resp,
		workspacePtrUint8, workspaceSizeUint8, numTransfers, resultChanUint8)
	go runTestUint16(readUint16Req, readUint16Resp, writeUint16Req, writeUint16Resp,
		workspacePtrUint16, workspaceSizeUint16, numTransfers, resultChanUint16)
	go runTestUint32(readUint32Req, readUint32Resp, writeUint32Req, writeUint32Resp,
		workspacePtrUint32, workspaceSizeUint32, numTransfers, resultChanUint32)
	go runTestUint64(readUint64Req, readUint64Resp, writeUint64Req, writeUint64Resp,
		workspacePtrUint64, workspaceSizeUint64, numTransfers, resultChanUint64)

	// Accumulate the test results.
	resultUint8 := <-resultChanUint8
	resultUint16 := <-resultChanUint16
	resultUint32 := <-resultChanUint32
	resultUint64 := <-resultChanUint64

	byteCount += uint64(resultUint8.byteCount)
	byteCount += uint64(resultUint16.byteCount)
	byteCount += uint64(resultUint32.byteCount)
	byteCount += uint64(resultUint64.byteCount)

	errorCount += uint64(resultUint8.errorCount)
	errorCount += uint64(resultUint16.errorCount)
	errorCount += uint64(resultUint32.errorCount)
	errorCount += uint64(resultUint64.errorCount)

	// Return the test results via shared memory.
	smi.WriteUInt64(writeResultReq, writeResultResp, byteCountPtr,
		smi.DefaultOptions, byteCount)
	smi.WriteUInt64(writeResultReq, writeResultResp, errorCountPtr,
		smi.DefaultOptions, errorCount)
}
