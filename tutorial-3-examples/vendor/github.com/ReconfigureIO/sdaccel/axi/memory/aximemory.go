//
// (c) 2017 ReconfigureIO
//
// <COPYRIGHT TERMS>
//

//
// AXI access interface to memory mapped RAM and I/O. This defines the memory
// access functions to support reading and writing of the various Go primitive
// types over the AXI bus. Note that in order to ensure the correct ordering of
// AXI channel requests and responses, each AXI client/server interface must
// only ever be accessed sequentially from within the same goroutine. A suitable
// memory arbitration component from the axi/protocol package will be required
// to support concurrent memory accesses.
//

/*

Package memory provides high level operations for working an AXI bus

*/
package memory

import (
	"github.com/ReconfigureIO/sdaccel/axi/protocol"
)

//
// Sets the maximum AXI burst length to use.
//
const maxAxiBurstSize = 64

//
// WriteUInt64 writes a single 64-bit unsigned data value to a word aligned
// address on the specified AXI memory bus, with the bottom three address bits
// being ignored. The status of the write transaction is returned as the boolean
// 'writeOk' flag.
//
func WriteUInt64(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	bufferedAccess bool,
	writeAddr uintptr,
	writeData uint64) bool {

	// Issue write request.
	go func() {
		clientAddr <- protocol.Addr{
			Addr:  writeAddr &^ uintptr(0x7),
			Size:  [3]bool{true, true, false},
			Burst: [2]bool{true, false},
			Cache: [4]bool{bufferedAccess, true, false, false}}
	}()

	// Perform full width 64-bit AXI write.
	writeStrobe := [8]bool{
		true, true, true, true, true, true, true, true}
	clientData <- protocol.WriteData{
		Data: writeData,
		Strb: writeStrobe,
		Last: true}
	writeResp := <-clientResp
	return !writeResp.Resp[1]
}

//
// ReadUInt64 reads a single 64-bit unsigned data value from a word aligned
// address on the specified AXI memory bus, with the bottom three address bits
// being ignored. TODO: The status of the read transaction should be returned
// as the boolean 'readOk' flag.
//
func ReadUInt64(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	bufferedAccess bool,
	readAddr uintptr) uint64 {

	// Issue read request.
	go func() {
		clientAddr <- protocol.Addr{
			Addr:  readAddr &^ uintptr(0x7),
			Size:  [3]bool{true, true, false},
			Burst: [2]bool{true, false},
			Cache: [4]bool{bufferedAccess, true, false, false}}
	}()

	// Process read response.
	readResp := <-clientData
	// TODO: return !readResp.Resp[1], readResp.Data
	return readResp.Data
}

//
// WriteUInt32 writes a single 32-bit unsigned data value to a word aligned
// address on the specified AXI memory bus, with the bottom two address bits
// being ignored. The status of the write transaction is returned as the boolean
// 'writeOk' flag.
//
func WriteUInt32(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	bufferedAccess bool,
	writeAddr uintptr,
	writeData uint32) bool {

	// Issue write request.
	go func() {
		clientAddr <- protocol.Addr{
			Addr:  writeAddr &^ uintptr(0x3),
			Size:  [3]bool{false, true, false},
			Burst: [2]bool{true, false},
			Cache: [4]bool{bufferedAccess, true, false, false}}
	}()

	// Map write data to appropriate byte lanes.
	var writeData64 uint64
	var writeStrobe [8]bool
	switch byte(writeAddr) & 0x4 {
	case 0x0:
		writeData64 = uint64(writeData)
		writeStrobe = [8]bool{
			true, true, true, true, false, false, false, false}
	default:
		writeData64 = uint64(writeData) << 32
		writeStrobe = [8]bool{
			false, false, false, false, true, true, true, true}
	}

	// Perform partial width 64-bit AXI write.
	clientData <- protocol.WriteData{
		Data: writeData64,
		Strb: writeStrobe,
		Last: true}
	writeResp := <-clientResp
	return !writeResp.Resp[1]
}

//
// ReadUInt32 reads a single 32-bit unsigned data value from a word aligned
// address on the specified AXI memory bus, with the bottom two address bits
// being ignored. TODO: The status of the read transaction should be returned as
// the boolean 'readOk' flag.
//
func ReadUInt32(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	bufferedAccess bool,
	readAddr uintptr) uint32 {

	// Issue read request.
	go func() {
		clientAddr <- protocol.Addr{
			Addr:  readAddr &^ uintptr(0x3),
			Size:  [3]bool{false, true, false},
			Burst: [2]bool{true, false},
			Cache: [4]bool{bufferedAccess, true, false, false}}
	}()

	// Select data from 64-bit read result.
	readResp := <-clientData
	var readData uint32
	switch byte(readAddr) & 0x4 {
	case 0x0:
		readData = uint32(readResp.Data)
	default:
		readData = uint32(readResp.Data >> 32)
	}
	// TODO: return !readResp.Resp[1], readData
	return readData
}

//
// WriteUInt16 writes a single 16-bit unsigned data value to a word aligned
// address on the specified AXI memory bus, with the bottom address bit being
// ignored. The status of the write transaction is returned as the boolean
// 'writeOk' flag.
//
func WriteUInt16(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	bufferedAccess bool,
	writeAddr uintptr,
	writeData uint16) bool {

	// Issue write request.
	go func() {
		clientAddr <- protocol.Addr{
			Addr:  writeAddr &^ uintptr(0x1),
			Size:  [3]bool{true, false, false},
			Burst: [2]bool{true, false},
			Cache: [4]bool{bufferedAccess, true, false, false}}
	}()

	// Map write data to appropriate byte lanes.
	var writeData64 uint64
	var writeStrobe [8]bool
	switch byte(writeAddr) & 0x6 {
	case 0x0:
		writeData64 = uint64(writeData)
		writeStrobe = [8]bool{
			true, true, false, false, false, false, false, false}
	case 0x2:
		writeData64 = uint64(writeData) << 16
		writeStrobe = [8]bool{
			false, false, true, true, false, false, false, false}
	case 0x4:
		writeData64 = uint64(writeData) << 32
		writeStrobe = [8]bool{
			false, false, false, false, true, true, false, false}
	default:
		writeData64 = uint64(writeData) << 48
		writeStrobe = [8]bool{
			false, false, false, false, false, false, true, true}
	}

	// Perform partial width 64-bit AXI write.
	clientData <- protocol.WriteData{
		Data: writeData64,
		Strb: writeStrobe,
		Last: true}
	writeResp := <-clientResp
	return !writeResp.Resp[1]
}

//
// ReadUInt16 reads a single 16-bit unsigned data value from a word aligned
// address on the specified AXI memory bus, with the bottom address bit being
// ignored. TODO: The status of the read transaction should be returned as the
// boolean 'readOk' flag.
//
func ReadUInt16(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	bufferedAccess bool,
	readAddr uintptr) uint16 {

	// Issue read request.
	go func() {
		clientAddr <- protocol.Addr{
			Addr:  readAddr &^ uintptr(0x1),
			Size:  [3]bool{true, false, false},
			Burst: [2]bool{true, false},
			Cache: [4]bool{bufferedAccess, true, false, false}}
	}()

	// Select data from 64-bit read result.
	readResp := <-clientData
	var readData uint16
	switch byte(readAddr) & 0x6 {
	case 0x0:
		readData = uint16(readResp.Data)
	case 0x2:
		readData = uint16(readResp.Data >> 16)
	case 0x4:
		readData = uint16(readResp.Data >> 32)
	default:
		readData = uint16(readResp.Data >> 48)
	}
	// TODO: return !readResp.Resp[1], readData
	return readData
}

//
// WriteUInt8 writes a single 8-bit unsigned data value to the specified AXI
// memory bus. The status of the write transaction is returned as the boolean
// 'writeOk' flag.
//
func WriteUInt8(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	bufferedAccess bool,
	writeAddr uintptr,
	writeData uint8) bool {

	// Issue write request.
	go func() {
		clientAddr <- protocol.Addr{
			Addr:  writeAddr,
			Size:  [3]bool{false, false, false},
			Burst: [2]bool{true, false},
			Cache: [4]bool{bufferedAccess, true, false, false}}
	}()

	// Map write data to appropriate byte lanes.
	var writeData64 uint64
	var writeStrobe [8]bool
	switch byte(writeAddr) & 0x7 {
	case 0x0:
		writeData64 = uint64(writeData)
		writeStrobe = [8]bool{
			true, false, false, false, false, false, false, false}
	case 0x1:
		writeData64 = uint64(writeData) << 8
		writeStrobe = [8]bool{
			false, true, false, false, false, false, false, false}
	case 0x2:
		writeData64 = uint64(writeData) << 16
		writeStrobe = [8]bool{
			false, false, true, false, false, false, false, false}
	case 0x3:
		writeData64 = uint64(writeData) << 24
		writeStrobe = [8]bool{
			false, false, false, true, false, false, false, false}
	case 0x4:
		writeData64 = uint64(writeData) << 32
		writeStrobe = [8]bool{
			false, false, false, false, true, false, false, false}
	case 0x5:
		writeData64 = uint64(writeData) << 40
		writeStrobe = [8]bool{
			false, false, false, false, false, true, false, false}
	case 0x6:
		writeData64 = uint64(writeData) << 48
		writeStrobe = [8]bool{
			false, false, false, false, false, false, true, false}
	default:
		writeData64 = uint64(writeData) << 56
		writeStrobe = [8]bool{
			false, false, false, false, false, false, false, true}
	}

	// Perform partial width 64-bit AXI write.
	clientData <- protocol.WriteData{
		Data: writeData64,
		Strb: writeStrobe,
		Last: true}
	writeResp := <-clientResp
	return !writeResp.Resp[1]
}

//
// ReadUInt8 reads a single 8-bit unsigned data value to the specified AXI
// memory bus. TODO: The status of the write transaction should be returned as
// the boolean 'readOk' flag.
//
func ReadUInt8(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	bufferedAccess bool,
	readAddr uintptr) uint8 {

	// Issue read request.
	go func() {
		clientAddr <- protocol.Addr{
			Addr:  readAddr,
			Size:  [3]bool{false, false, false},
			Burst: [2]bool{true, false},
			Cache: [4]bool{bufferedAccess, true, false, false}}
	}()

	// Select data from 64-bit read result.
	readResp := <-clientData
	var readData uint8
	switch byte(readAddr) & 0x7 {
	case 0x0:
		readData = uint8(readResp.Data)
	case 0x1:
		readData = uint8(readResp.Data >> 8)
	case 0x2:
		readData = uint8(readResp.Data >> 16)
	case 0x3:
		readData = uint8(readResp.Data >> 24)
	case 0x4:
		readData = uint8(readResp.Data >> 32)
	case 0x5:
		readData = uint8(readResp.Data >> 40)
	case 0x6:
		readData = uint8(readResp.Data >> 48)
	default:
		readData = uint8(readResp.Data >> 56)
	}
	// TODO: return !readResp.Resp[1], readData
	return readData
}

//
// WriteBurstUInt64 writes an incrementing burst of 64-bit unsigned data values
// to a word aligned address on the specified AXI memory bus, with the bottom
// three address bits being ignored. The status of the write transaction is
// returned as the boolean 'burstOk' flag.
//
func WriteBurstUInt64(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	bufferedAccess bool,
	writeAddr uintptr,
	writeLength uint32,
	writeDataChan <-chan uint64) bool {

	// Get aligned address.
	alignedAddr := writeAddr &^ uintptr(0x7)

	// Divide the transaction into burst sequences.
	burstSize := byte(maxAxiBurstSize)
	burstOk := true
	for writeLength != 0 {
		if writeLength < maxAxiBurstSize {
			burstSize = byte(writeLength)
		}

		// Perform full width 64-bit AXI burst writes.
		go func() {
			clientAddr <- protocol.Addr{
				Addr:  alignedAddr,
				Len:   burstSize - 1,
				Size:  [3]bool{true, true, false},
				Burst: [2]bool{true, false},
				Cache: [4]bool{bufferedAccess, true, false, false}}
		}()

		// Loops over the required number of burst transactions.
		for i := burstSize; i != 0; i-- {
			writeData := <-writeDataChan
			clientData <- protocol.WriteData{
				Data: writeData,
				Strb: [8]bool{
					true, true, true, true,
					true, true, true, true},
				Last: i == 1}
		}

		// Update the burst counter and status flag.
		writeResp := <-clientResp
		burstOk = burstOk && !writeResp.Resp[1]
		writeLength -= uint32(burstSize)
		alignedAddr += uintptr(burstSize) << 3
	}
	return burstOk
}

//
// ReadBurstUInt64 reads an incrementing burst of 64-bit unsigned data values
// from a word aligned address on the specified AXI memory bus, with the bottom
// three address bits being ignored. The status of the read transaction is
// returned as the boolean 'burstOk' flag.
//
func ReadBurstUInt64(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	bufferedAccess bool,
	readAddr uintptr,
	readLength uint32,
	readDataChan chan<- uint64) bool {

	// Divide the transaction into burst sequences.
	alignedAddr := readAddr &^ uintptr(0x7)
	burstSize := byte(maxAxiBurstSize)
	burstOk := true
	for readLength != 0 {
		if readLength < maxAxiBurstSize {
			burstSize = byte(readLength)
		}

		// Perform full width 64-bit AXI burst reads.
		go func() {
			clientAddr <- protocol.Addr{
				Addr:  alignedAddr,
				Len:   burstSize - 1,
				Size:  [3]bool{true, true, false},
				Burst: [2]bool{true, false},
				Cache: [4]bool{bufferedAccess, true, false, false}}
		}()

		// Loops until read data contains 'last' flag. Only the final
		// burst status is of interest.
		getNext := true
		for getNext {
			readData := <-clientData
			readDataChan <- readData.Data
			if readData.Last {
				burstOk = burstOk && !readData.Resp[1]
			}
			getNext = !readData.Last
		}

		// Update the burst counter and status flag.
		readLength -= uint32(burstSize)
		alignedAddr += uintptr(burstSize) << 3
	}
	return burstOk
}

//
// WriteBurstUInt32 writes an incrementing burst of 32-bit unsigned data values
// to a word aligned address on the specified AXI memory bus, with the bottom
// two address bits being ignored. The status of the write transaction is
// returned as the boolean 'burstOk' flag.
//
func WriteBurstUInt32(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	bufferedAccess bool,
	writeAddr uintptr,
	writeLength uint32,
	writeDataChan <-chan uint32) bool {

	// Get aligned address and initial strobe phase.
	alignedAddr := writeAddr &^ uintptr(0x3)
	strobePhase := byte(writeAddr)
	var writeData64 uint64
	var writeStrobe [8]bool

	// Divide the transaction into burst sequences.
	burstSize := byte(maxAxiBurstSize)
	burstOk := true
	for writeLength != 0 {
		if writeLength < maxAxiBurstSize {
			burstSize = byte(writeLength)
		}

		// Perform partial width AXI burst writes.
		go func() {
			clientAddr <- protocol.Addr{
				Addr:  alignedAddr,
				Len:   burstSize - 1,
				Size:  [3]bool{false, true, false},
				Burst: [2]bool{true, false},
				Cache: [4]bool{bufferedAccess, true, false, false}}
		}()

		// Loops over the required number of burst transactions.
		for i := burstSize; i != 0; i-- {
			writeData := <-writeDataChan

			// Map write data to appropriate byte lanes.
			switch strobePhase & 0x4 {
			case 0x0:
				writeData64 = uint64(writeData)
				writeStrobe = [8]bool{
					true, true, true, true, false, false, false, false}
			default:
				writeData64 = uint64(writeData) << 32
				writeStrobe = [8]bool{
					false, false, false, false, true, true, true, true}
			}

			// Perform partial width 64-bit AXI write.
			clientData <- protocol.WriteData{
				Data: writeData64,
				Strb: writeStrobe,
				Last: i == 1}
			strobePhase += 0x4
		}

		// Update the burst counter and status flag.
		writeResp := <-clientResp
		burstOk = burstOk && !writeResp.Resp[1]
		writeLength -= uint32(burstSize)
		alignedAddr += uintptr(burstSize) << 2
	}
	return burstOk
}

//
// ReadBurstUInt32 reads an incrementing burst of 32-bit unsigned data values
// from a word aligned address on the specified AXI memory bus, with the bottom
// two address bits being ignored. The status of the read transaction is
// returned as the boolean 'burstOk' flag.
//
func ReadBurstUInt32(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	bufferedAccess bool,
	readAddr uintptr,
	readLength uint32,
	readDataChan chan<- uint32) bool {

	// Get aligned address and initial read phase.
	alignedAddr := readAddr &^ uintptr(0x3)
	readPhase := byte(readAddr)

	// Divide the transaction into burst sequences.
	burstSize := byte(maxAxiBurstSize)
	burstOk := true
	for readLength != 0 {
		if readLength < maxAxiBurstSize {
			burstSize = byte(readLength)
		}

		// Perform partial width AXI burst writes.
		go func() {
			clientAddr <- protocol.Addr{
				Addr:  alignedAddr,
				Len:   burstSize - 1,
				Size:  [3]bool{false, true, false},
				Burst: [2]bool{true, false},
				Cache: [4]bool{bufferedAccess, true, false, false}}
		}()

		// Loops until read data contains 'last' flag. Only the final
		// burst status is of interest.
		getNext := true
		for getNext {
			readData := <-clientData
			var dataVal uint32
			switch readPhase & 0x4 {
			case 0x0:
				dataVal = uint32(readData.Data)
			default:
				dataVal = uint32(readData.Data >> 32)
			}
			readDataChan <- dataVal
			if readData.Last {
				burstOk = burstOk && !readData.Resp[1]
			}
			readPhase += 0x4
			getNext = !readData.Last
		}

		// Update the burst counter and status flag.
		readLength -= uint32(burstSize)
		alignedAddr += uintptr(burstSize) << 2
	}
	return burstOk
}

//
// WriteBurstUInt16 writes an incrementing burst of 16-bit unsigned data values
// to a word aligned address on the specified AXI memory bus, with the bottom
// address bit being ignored. The status of the write transaction is returned
// as the boolean 'burstOk' flag.
//
func WriteBurstUInt16(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	bufferedAccess bool,
	writeAddr uintptr,
	writeLength uint32,
	writeDataChan <-chan uint16) bool {

	// Get aligned address and initial strobe phase.
	alignedAddr := writeAddr &^ uintptr(0x1)
	strobePhase := byte(writeAddr)
	var writeData64 uint64
	var writeStrobe [8]bool

	// Divide the transaction into burst sequences.
	burstSize := byte(maxAxiBurstSize)
	burstOk := true
	for writeLength != 0 {
		if writeLength < maxAxiBurstSize {
			burstSize = byte(writeLength)
		}

		// Perform partial width AXI burst writes.
		go func() {
			clientAddr <- protocol.Addr{
				Addr:  alignedAddr,
				Len:   burstSize - 1,
				Size:  [3]bool{true, false, false},
				Burst: [2]bool{true, false},
				Cache: [4]bool{bufferedAccess, true, false, false}}
		}()

		// Loops over the required number of burst transactions.
		for i := burstSize; i != 0; i-- {
			writeData := <-writeDataChan

			// Map write data to appropriate byte lanes.
			switch strobePhase & 0x6 {
			case 0x0:
				writeData64 = uint64(writeData)
				writeStrobe = [8]bool{
					true, true, false, false, false, false, false, false}
			case 0x2:
				writeData64 = uint64(writeData) << 16
				writeStrobe = [8]bool{
					false, false, true, true, false, false, false, false}
			case 0x4:
				writeData64 = uint64(writeData) << 32
				writeStrobe = [8]bool{
					false, false, false, false, true, true, false, false}
			default:
				writeData64 = uint64(writeData) << 48
				writeStrobe = [8]bool{
					false, false, false, false, false, false, true, true}
			}

			// Perform partial width 64-bit AXI write.
			clientData <- protocol.WriteData{
				Data: writeData64,
				Strb: writeStrobe,
				Last: i == 1}
			strobePhase += 0x2
		}

		// Update the burst counter and status flag.
		writeResp := <-clientResp
		burstOk = burstOk && !writeResp.Resp[1]
		writeLength -= uint32(burstSize)
		alignedAddr += uintptr(burstSize) << 1
	}
	return burstOk
}

//
// ReadBurstUInt16 reads an incrementing burst of 16-bit unsigned data values
// from a word aligned address on the specified AXI memory bus, with the bottom
// address bit being ignored. The status of the read transaction is returned as
// the boolean 'burstOk' flag.
//
func ReadBurstUInt16(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	bufferedAccess bool,
	readAddr uintptr,
	readLength uint32,
	readDataChan chan<- uint16) bool {

	// Get aligned address and initial read phase.
	alignedAddr := readAddr &^ uintptr(0x1)
	readPhase := byte(readAddr)

	// Divide the transaction into burst sequences.
	burstSize := byte(maxAxiBurstSize)
	burstOk := true
	for readLength != 0 {
		if readLength < maxAxiBurstSize {
			burstSize = byte(readLength)
		}

		// Perform partial width AXI burst writes.
		go func() {
			clientAddr <- protocol.Addr{
				Addr:  alignedAddr,
				Len:   burstSize - 1,
				Size:  [3]bool{true, false, false},
				Burst: [2]bool{true, false},
				Cache: [4]bool{bufferedAccess, true, false, false}}
		}()

		// Loops until read data contains 'last' flag. Only the final
		// burst status is of interest.
		getNext := true
		for getNext {
			readData := <-clientData
			switch readPhase & 0x6 {
			case 0x0:
				readDataChan <- uint16(readData.Data)
			case 0x2:
				readDataChan <- uint16(readData.Data >> 16)
			case 0x4:
				readDataChan <- uint16(readData.Data >> 32)
			default:
				readDataChan <- uint16(readData.Data >> 48)
			}
			if readData.Last {
				burstOk = burstOk && !readData.Resp[1]
			}
			readPhase += 0x2
			getNext = !readData.Last
		}

		// Update the burst counter and status flag.
		readLength -= uint32(burstSize)
		alignedAddr += uintptr(burstSize) << 1
	}
	return burstOk
}

//
// WriteBurstUInt8 writes an incrementing burst of 8-bit unsigned data values
// on the specified AXI memory bus. The status of the write transaction is
// returned as the boolean 'burstOk' flag.
//
func WriteBurstUInt8(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	bufferedAccess bool,
	writeAddr uintptr,
	writeLength uint32,
	writeDataChan <-chan uint8) bool {

	// Get aligned address and initial strobe phase.
	alignedAddr := writeAddr
	strobePhase := byte(writeAddr)
	var writeData64 uint64
	var writeStrobe [8]bool

	// Divide the transaction into burst sequences.
	burstSize := byte(maxAxiBurstSize)
	burstOk := true
	for writeLength != 0 {
		if writeLength < maxAxiBurstSize {
			burstSize = byte(writeLength)
		}

		// Perform partial width AXI burst writes.
		go func() {
			clientAddr <- protocol.Addr{
				Addr:  alignedAddr,
				Len:   burstSize - 1,
				Size:  [3]bool{false, false, false},
				Burst: [2]bool{true, false},
				Cache: [4]bool{bufferedAccess, true, false, false}}
		}()

		// Loops over the required number of burst transactions.
		for i := burstSize; i != 0; i-- {
			writeData := <-writeDataChan

			// Map write data to appropriate byte lanes.
			switch strobePhase & 0x7 {
			case 0x0:
				writeData64 = uint64(writeData)
				writeStrobe = [8]bool{
					true, false, false, false, false, false, false, false}
			case 0x1:
				writeData64 = uint64(writeData) << 8
				writeStrobe = [8]bool{
					false, true, false, false, false, false, false, false}
			case 0x2:
				writeData64 = uint64(writeData) << 16
				writeStrobe = [8]bool{
					false, false, true, false, false, false, false, false}
			case 0x3:
				writeData64 = uint64(writeData) << 24
				writeStrobe = [8]bool{
					false, false, false, true, false, false, false, false}
			case 0x4:
				writeData64 = uint64(writeData) << 32
				writeStrobe = [8]bool{
					false, false, false, false, true, false, false, false}
			case 0x5:
				writeData64 = uint64(writeData) << 40
				writeStrobe = [8]bool{
					false, false, false, false, false, true, false, false}
			case 0x6:
				writeData64 = uint64(writeData) << 48
				writeStrobe = [8]bool{
					false, false, false, false, false, false, true, false}
			default:
				writeData64 = uint64(writeData) << 56
				writeStrobe = [8]bool{
					false, false, false, false, false, false, false, true}
			}

			// Perform partial width 64-bit AXI write.
			clientData <- protocol.WriteData{
				Data: writeData64,
				Strb: writeStrobe,
				Last: i == 1}
			strobePhase += 0x1
		}

		// Update the burst counter and status flag.
		writeResp := <-clientResp
		burstOk = burstOk && !writeResp.Resp[1]
		writeLength -= uint32(burstSize)
		alignedAddr += uintptr(burstSize)
	}
	return burstOk
}

//
// ReadBurstUInt8 reads an incrementing burst of 8-bit unsigned data values
// from a word aligned address on the specified AXI memory bus, with the bottom
// address bit being ignored. The status of the read transaction is returned as
// the boolean 'burstOk' flag.
//
func ReadBurstUInt8(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	bufferedAccess bool,
	readAddr uintptr,
	readLength uint32,
	readDataChan chan<- uint8) bool {

	// Get aligned address and initial read phase.
	alignedAddr := readAddr
	readPhase := byte(readAddr)

	// Divide the transaction into burst sequences.
	burstSize := byte(maxAxiBurstSize)
	burstOk := true
	for readLength != 0 {
		if readLength < maxAxiBurstSize {
			burstSize = byte(readLength)
		}

		// Perform partial width AXI burst writes.
		go func() {
			clientAddr <- protocol.Addr{
				Addr:  alignedAddr,
				Len:   burstSize - 1,
				Size:  [3]bool{false, false, false},
				Burst: [2]bool{true, false},
				Cache: [4]bool{bufferedAccess, true, false, false}}
		}()

		// Loops until read data contains 'last' flag. Only the final
		// burst status is of interest.
		getNext := true
		for getNext {
			readData := <-clientData
			switch readPhase & 0x7 {
			case 0x0:
				readDataChan <- uint8(readData.Data)
			case 0x1:
				readDataChan <- uint8(readData.Data >> 8)
			case 0x2:
				readDataChan <- uint8(readData.Data >> 16)
			case 0x3:
				readDataChan <- uint8(readData.Data >> 24)
			case 0x4:
				readDataChan <- uint8(readData.Data >> 32)
			case 0x5:
				readDataChan <- uint8(readData.Data >> 40)
			case 0x6:
				readDataChan <- uint8(readData.Data >> 48)
			default:
				readDataChan <- uint8(readData.Data >> 56)
			}
			if readData.Last {
				burstOk = burstOk && !readData.Resp[1]
			}
			readPhase += 0x1
			getNext = !readData.Last
		}

		// Update the burst counter and status flag.
		readLength -= uint32(burstSize)
		alignedAddr += uintptr(burstSize)
	}
	return burstOk
}
