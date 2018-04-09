//
// (c) 2018 ReconfigureIO
//
// <COPYRIGHT TERMS>
//

//
// Package smi/protocol provides low level primitives and data types for working
// with the SMI protocol.
//
package smi

//
// Constants specifying the supported SMI frame type bytes.
//
const (
	SmiMemWriteReq  = 0x01 // SMI memory write request.
	SmiMemWriteResp = 0xFE // SMI memory write response.
	SmiMemReadReq   = 0x02 // SMI memory read request.
	SmiMemReadResp  = 0xFD // SMI memory read response.
)

//
// Default options constant
//
const (
	DefaultOptions = uint8(0x00) // Use default buffered read or write.
)

//
// Constants specifying the supported SMI memory access options.
//
const (
	MemOptUnbuffered = uint8(0x01) // Perform direct unbuffered read or write.
)

//
// Specify the standard burst fragment size as an integer number of bytes.
//
const SmiMemBurstSize = 256

//
// The maximum frame size is derived from the SmiMemBurstSize parameter
// and can contain the specified amount of data plus up to 16 bytes of
// header information.
//
const SmiMemFrame64Size = 2 + SmiMemBurstSize/8

//
// Specify the number of in-flight transactions supported by each
// arbitrated SMI port.
//
const SmiMemInFlightLimit = 4

//
// Type Flit64 specifies an SMI flit format with a 64-bit datapath.
//
type Flit64 struct {
	Data [8]uint8
	Eofc uint8
}

//
// Forwards a single Flit64 based SMI frame from an input channel to an output
// channel with intermediate buffering. The buffer has capacity to store a
// complete frame, with data being available at the output as soon as it has
// been received on the input.
// TODO: Update once there is a fix for the channel size compiler limitation.
//
func ForwardFrame64(
	smiInput <-chan Flit64,
	smiOutput chan<- Flit64) {
	smiBuffer := make(chan Flit64, 34 /* SmiMemFrame64Size */)

	go func() {
		hasNextInputFlit := true
		for hasNextInputFlit {
			inputFlitData := <-smiInput
			smiBuffer <- inputFlitData
			hasNextInputFlit = inputFlitData.Eofc == uint8(0)
		}
	}()

	hasNextOutputFlit := true
	for hasNextOutputFlit {
		outputFlitData := <-smiBuffer
		smiOutput <- outputFlitData
		hasNextOutputFlit = outputFlitData.Eofc == uint8(0)
	}
}

//
// Assembles a single Flit64 based SMI frame from an input channel, copying the
// frame to the output channel once the entire frame has been received. The
// maximum frame size is derived from the SmiMemBurstSize parameter and can
// contain the specified amount of payload data plus up to 16 bytes of header
// information.
// TODO: Update once there is a fix for the channel size compiler limitation.
//
func AssembleFrame64(
	smiInput <-chan Flit64,
	smiOutput chan<- Flit64) {
	smiBuffer := make(chan Flit64, 34 /* SmiMemFrame64Size */)

	hasNextInputFlit := true
	for hasNextInputFlit {
		inputFlitData := <-smiInput
		smiBuffer <- inputFlitData
		hasNextInputFlit = inputFlitData.Eofc == uint8(0)
	}

	hasNextOutputFlit := true
	for hasNextOutputFlit {
		outputFlitData := <-smiBuffer
		smiOutput <- outputFlitData
		hasNextOutputFlit = outputFlitData.Eofc == uint8(0)
	}
}

//
// Package arbitrate provides reusable arbitrators for SMI transactions.
//

//
// manageUpstreamPort provides transaction management for the arbitrated
// upstream ports. This includes header tag switching to allow request and
// response message pairs to be matched up.
//
func manageUpstreamPort(
	upstreamRequest <-chan Flit64,
	upstreamResponse chan<- Flit64,
	taggedRequest chan<- Flit64,
	taggedResponse <-chan Flit64,
	transferReq chan<- uint8,
	portId uint8) {

	// Split the tags into upper and lower bytes for efficient access.
	// TODO: The array and channel sizes here should be set using the
	// SmiMemInFlightLimit constant once supported by the compiler.
	var tagTableLower [4]uint8
	var tagTableUpper [4]uint8
	tagFifo := make(chan uint8, 4)

	// Set up the local tag values.
	for tagInit := uint8(0); tagInit != 4; tagInit++ {
		tagFifo <- tagInit
	}

	// Start goroutine for tag replacement on requests.
	go func() {
		for {

			// Do tag replacement on header.
			headerFlit := <-upstreamRequest
			tagId := <-tagFifo
			tagTableLower[tagId] = headerFlit.Data[2]
			tagTableUpper[tagId] = headerFlit.Data[3]
			headerFlit.Data[2] = portId
			headerFlit.Data[3] = tagId
			transferReq <- portId
			taggedRequest <- headerFlit

			// Copy remaining flits from upstream to downstream.
			moreFlits := headerFlit.Eofc == 0
			for moreFlits {
				bodyFlit := <-upstreamRequest
				moreFlits = bodyFlit.Eofc == 0
				taggedRequest <- bodyFlit
			}
		}
	}()

	// Carry out tag replacement on responses.
	for {

		// Extract tag ID from header and use it to look up replacement.
		headerFlit := <-taggedResponse
		tagId := headerFlit.Data[3]
		headerFlit.Data[2] = tagTableLower[tagId]
		headerFlit.Data[3] = tagTableUpper[tagId]
		tagFifo <- tagId
		upstreamResponse <- headerFlit

		// Copy remaining flits from downstream to upstream.
		moreFlits := headerFlit.Eofc == 0
		for moreFlits {
			bodyFlit := <-taggedResponse
			moreFlits = bodyFlit.Eofc == 0
			upstreamResponse <- bodyFlit
		}
	}
}

//
// ArbitrateX2 is a goroutine for providing arbitration between two pairs of
// SMI request/response channels. This uses tag matching and substitution on
// bytes 2 and 3 of each transfer to ensure that response frames are correctly
// routed to the source of the original request.
//
func ArbitrateX2(
	upstreamRequestA <-chan Flit64,
	upstreamResponseA chan<- Flit64,
	upstreamRequestB <-chan Flit64,
	upstreamResponseB chan<- Flit64,
	downstreamRequest chan<- Flit64,
	downstreamResponse <-chan Flit64) {

	// Define local channel connections.
	taggedRequestA := make(chan Flit64, 1)
	taggedResponseA := make(chan Flit64, 1)
	taggedRequestB := make(chan Flit64, 1)
	taggedResponseB := make(chan Flit64, 1)
	transferReqA := make(chan uint8, 1)
	transferReqB := make(chan uint8, 1)

	// Run the upstream port management routines.
	go manageUpstreamPort(upstreamRequestA, upstreamResponseA,
		taggedRequestA, taggedResponseA, transferReqA, uint8(1))
	go manageUpstreamPort(upstreamRequestB, upstreamResponseB,
		taggedRequestB, taggedResponseB, transferReqB, uint8(2))

	// Arbitrate between transfer requests.
	go func() {
		for {

			// Gets port ID of active input.
			var portId uint8
			select {
			case portId = <-transferReqA:
			case portId = <-transferReqB:
			}

			// Copy over input data.
			var reqFlit Flit64
			moreFlits := true
			for moreFlits {
				switch portId {
				case 1:
					reqFlit = <-taggedRequestA
				default:
					reqFlit = <-taggedRequestB
				}
				downstreamRequest <- reqFlit
				moreFlits = reqFlit.Eofc == 0
			}
		}
	}()

	// Steer transfer responses.
	portId := uint8(0)
	isHeaderFlit := true
	for {
		respFlit := <-downstreamResponse
		if isHeaderFlit {
			portId = respFlit.Data[2]
		}
		switch portId {
		case 1:
			taggedResponseA <- respFlit
		case 2:
			taggedResponseB <- respFlit
		default:
			// Discard invalid flit.
		}
		isHeaderFlit = respFlit.Eofc != 0
	}
}

//
// ArbitrateX3 is a goroutine for providing arbitration between three pairs of
// SMI request/response channels. This uses tag matching and substitution on
// bytes 2 and 3 of each transfer to ensure that response frames are correctly
// routed to the source of the original request.
//
func ArbitrateX3(
	upstreamRequestA <-chan Flit64,
	upstreamResponseA chan<- Flit64,
	upstreamRequestB <-chan Flit64,
	upstreamResponseB chan<- Flit64,
	upstreamRequestC <-chan Flit64,
	upstreamResponseC chan<- Flit64,
	downstreamRequest chan<- Flit64,
	downstreamResponse <-chan Flit64) {

	// Define local channel connections.
	taggedRequestA := make(chan Flit64, 1)
	taggedResponseA := make(chan Flit64, 1)
	taggedRequestB := make(chan Flit64, 1)
	taggedResponseB := make(chan Flit64, 1)
	taggedRequestC := make(chan Flit64, 1)
	taggedResponseC := make(chan Flit64, 1)
	transferReqA := make(chan uint8, 1)
	transferReqB := make(chan uint8, 1)
	transferReqC := make(chan uint8, 1)

	// Run the upstream port management routines.
	go manageUpstreamPort(upstreamRequestA, upstreamResponseA,
		taggedRequestA, taggedResponseA, transferReqA, uint8(1))
	go manageUpstreamPort(upstreamRequestB, upstreamResponseB,
		taggedRequestB, taggedResponseB, transferReqB, uint8(2))
	go manageUpstreamPort(upstreamRequestC, upstreamResponseC,
		taggedRequestC, taggedResponseC, transferReqC, uint8(3))

	// Arbitrate between transfer requests.
	go func() {
		for {

			// Gets port ID of active input.
			var portId uint8
			select {
			case portId = <-transferReqA:
			case portId = <-transferReqB:
			case portId = <-transferReqC:
			}

			// Copy over input data.
			var reqFlit Flit64
			moreFlits := true
			for moreFlits {
				switch portId {
				case 1:
					reqFlit = <-taggedRequestA
				case 2:
					reqFlit = <-taggedRequestB
				default:
					reqFlit = <-taggedRequestC
				}
				downstreamRequest <- reqFlit
				moreFlits = reqFlit.Eofc == 0
			}
		}
	}()

	// Steer transfer responses.
	portId := uint8(0)
	isHeaderFlit := true
	for {
		respFlit := <-downstreamResponse
		if isHeaderFlit {
			portId = respFlit.Data[2]
		}
		switch portId {
		case 1:
			taggedResponseA <- respFlit
		case 2:
			taggedResponseB <- respFlit
		case 3:
			taggedResponseC <- respFlit
		default:
			// Discard invalid flit.
		}
		isHeaderFlit = respFlit.Eofc != 0
	}
}

//
// ArbitrateX4 is a goroutine for providing arbitration between four pairs of
// SMI request/response channels. This uses tag matching and substitution on
// bytes 2 and 3 of each transfer to ensure that response frames are correctly
// routed to the source of the original request.
//
func ArbitrateX4(
	upstreamRequestA <-chan Flit64,
	upstreamResponseA chan<- Flit64,
	upstreamRequestB <-chan Flit64,
	upstreamResponseB chan<- Flit64,
	upstreamRequestC <-chan Flit64,
	upstreamResponseC chan<- Flit64,
	upstreamRequestD <-chan Flit64,
	upstreamResponseD chan<- Flit64,
	downstreamRequest chan<- Flit64,
	downstreamResponse <-chan Flit64) {

	// Define local channel connections.
	taggedRequestA := make(chan Flit64, 1)
	taggedResponseA := make(chan Flit64, 1)
	taggedRequestB := make(chan Flit64, 1)
	taggedResponseB := make(chan Flit64, 1)
	taggedRequestC := make(chan Flit64, 1)
	taggedResponseC := make(chan Flit64, 1)
	taggedRequestD := make(chan Flit64, 1)
	taggedResponseD := make(chan Flit64, 1)
	transferReqA := make(chan uint8, 1)
	transferReqB := make(chan uint8, 1)
	transferReqC := make(chan uint8, 1)
	transferReqD := make(chan uint8, 1)

	// Run the upstream port management routines.
	go manageUpstreamPort(upstreamRequestA, upstreamResponseA,
		taggedRequestA, taggedResponseA, transferReqA, uint8(1))
	go manageUpstreamPort(upstreamRequestB, upstreamResponseB,
		taggedRequestB, taggedResponseB, transferReqB, uint8(2))
	go manageUpstreamPort(upstreamRequestC, upstreamResponseC,
		taggedRequestC, taggedResponseC, transferReqC, uint8(3))
	go manageUpstreamPort(upstreamRequestD, upstreamResponseD,
		taggedRequestD, taggedResponseD, transferReqD, uint8(4))

	// Arbitrate between transfer requests.
	go func() {
		for {

			// Gets port ID of active input.
			var portId uint8
			select {
			case portId = <-transferReqA:
			case portId = <-transferReqB:
			case portId = <-transferReqC:
			case portId = <-transferReqD:
			}

			// Copy over input data.
			var reqFlit Flit64
			moreFlits := true
			for moreFlits {
				switch portId {
				case 1:
					reqFlit = <-taggedRequestA
				case 2:
					reqFlit = <-taggedRequestB
				case 3:
					reqFlit = <-taggedRequestC
				default:
					reqFlit = <-taggedRequestD
				}
				downstreamRequest <- reqFlit
				moreFlits = reqFlit.Eofc == 0
			}
		}
	}()

	// Steer transfer responses.
	portId := uint8(0)
	isHeaderFlit := true
	for {
		respFlit := <-downstreamResponse
		if isHeaderFlit {
			portId = respFlit.Data[2]
		}
		switch portId {
		case 1:
			taggedResponseA <- respFlit
		case 2:
			taggedResponseB <- respFlit
		case 3:
			taggedResponseC <- respFlit
		case 4:
			taggedResponseD <- respFlit
		default:
			// Discard invalid flit.
		}
		isHeaderFlit = respFlit.Eofc != 0
	}
}

//
// Package smi/memory provides high level operations for SMI access to memory
// mapped RAM and I/O. This defines the memory access functions to support
// reading and writing of the various Go primitive types over an SMI memory
// access endpoint. Note that in order to ensure the correct ordering of SMI
// channel requests and responses, each SMI client/server interface must only
// ever be accessed sequentially from within the same goroutine. A suitable
// memory arbitration component from the smi/protocol package will be required
// to support concurrent memory accesses on a single SMI memory access
// endpoint.
//

//
// WriteUInt64 writes a single 64-bit unsigned data value to a word aligned
// address on the specified SMI memory endpoint, with the bottom three address
// bits being ignored. The status of the write transaction is returned as the
// boolean 'writeOk' flag.
//
func WriteUInt64(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddr uintptr,
	writeOptions uint8,
	writeData uint64) bool {

	// Assemble the request message.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemWriteReq),
			uint8(writeOptions),
			uint8(0),
			uint8(0),
			uint8(writeAddr) & 0xF8,
			uint8(writeAddr >> 8),
			uint8(writeAddr >> 16),
			uint8(writeAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(writeAddr >> 32),
			uint8(writeAddr >> 40),
			uint8(writeAddr >> 48),
			uint8(writeAddr >> 56),
			uint8(8),
			uint8(0),
			uint8(writeData),
			uint8(writeData >> 8)}}

	reqFlit3 := Flit64{
		Eofc: 6,
		Data: [8]uint8{
			uint8(writeData >> 16),
			uint8(writeData >> 24),
			uint8(writeData >> 32),
			uint8(writeData >> 40),
			uint8(writeData >> 48),
			uint8(writeData >> 56),
			uint8(0),
			uint8(0)}}

	// Transmit the request message.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2
	smiRequest <- reqFlit3

	// Accept the response message.
	respFlit := <-smiResponse
	var writeOk bool
	if (respFlit.Data[1] & 0x02) == uint8(0x00) {
		writeOk = true
	} else {
		writeOk = false
	}
	return writeOk
}

//
// WriteUInt32 writes a single 32-bit unsigned data value to a word aligned
// address on the specified SMI memory endpoint, with the bottom two address
// bits being ignored. The status of the write transaction is returned as the
// boolean 'writeOk' flag.
//
func WriteUInt32(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddr uintptr,
	writeOptions uint8,
	writeData uint32) bool {

	// Assemble the request message.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemWriteReq),
			uint8(writeOptions),
			uint8(0),
			uint8(0),
			uint8(writeAddr) & 0xFC,
			uint8(writeAddr >> 8),
			uint8(writeAddr >> 16),
			uint8(writeAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(writeAddr >> 32),
			uint8(writeAddr >> 40),
			uint8(writeAddr >> 48),
			uint8(writeAddr >> 56),
			uint8(4),
			uint8(0),
			uint8(writeData),
			uint8(writeData >> 8)}}

	reqFlit3 := Flit64{
		Eofc: 2,
		Data: [8]uint8{
			uint8(writeData >> 16),
			uint8(writeData >> 24),
			uint8(0),
			uint8(0),
			uint8(0),
			uint8(0),
			uint8(0),
			uint8(0)}}

	// Transmit the request message.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2
	smiRequest <- reqFlit3

	// Accept the response message.
	respFlit := <-smiResponse
	var writeOk bool
	if (respFlit.Data[1] & 0x02) == uint8(0x00) {
		writeOk = true
	} else {
		writeOk = false
	}
	return writeOk
}

//
// WriteUInt16 writes a single 16-bit unsigned data value to a word aligned
// address on the specified SMI memory endpoint, with the bottom address
// bit being ignored. The status of the write transaction is returned as the
// boolean 'writeOk' flag.
//
func WriteUInt16(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddr uintptr,
	writeOptions uint8,
	writeData uint16) bool {

	// Assemble the request message.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemWriteReq),
			uint8(writeOptions),
			uint8(0),
			uint8(0),
			uint8(writeAddr) & 0xFE,
			uint8(writeAddr >> 8),
			uint8(writeAddr >> 16),
			uint8(writeAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 8,
		Data: [8]uint8{
			uint8(writeAddr >> 32),
			uint8(writeAddr >> 40),
			uint8(writeAddr >> 48),
			uint8(writeAddr >> 56),
			uint8(2),
			uint8(0),
			uint8(writeData),
			uint8(writeData >> 8)}}

	// Transmit the request message.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Accept the response message.
	respFlit := <-smiResponse
	var writeOk bool
	if (respFlit.Data[1] & 0x02) == uint8(0x00) {
		writeOk = true
	} else {
		writeOk = false
	}
	return writeOk
}

//
// WriteUInt8 writes a single 8-bit unsigned data value to a byte aligned
// address on the specified SMI memory endpoint. The status of the write
// transaction is returned as the boolean 'writeOk' flag.
//
func WriteUInt8(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddr uintptr,
	writeOptions uint8,
	writeData uint8) bool {

	// Assemble the request message.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemWriteReq),
			uint8(writeOptions),
			uint8(0),
			uint8(0),
			uint8(writeAddr),
			uint8(writeAddr >> 8),
			uint8(writeAddr >> 16),
			uint8(writeAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 7,
		Data: [8]uint8{
			uint8(writeAddr >> 32),
			uint8(writeAddr >> 40),
			uint8(writeAddr >> 48),
			uint8(writeAddr >> 56),
			uint8(1),
			uint8(0),
			uint8(writeData),
			uint8(0)}}

	// Transmit the request message.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Accept the response message.
	respFlit := <-smiResponse
	var writeOk bool
	if (respFlit.Data[1] & 0x02) == uint8(0x00) {
		writeOk = true
	} else {
		writeOk = false
	}
	return writeOk
}

//
// ReadUInt64 reads a single 64-bit unsigned data value from a word aligned
// address on the specified SMI memory endpoint, with the bottom three address
// bits being ignored.
// TODO: The status of the write transaction should also be returned as the
// boolean 'readOk' flag.
//
func ReadUInt64(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddr uintptr,
	readOptions uint8) uint64 {

	// Assemble the request message.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemReadReq),
			uint8(readOptions),
			uint8(0),
			uint8(0),
			uint8(readAddr) & 0xF8,
			uint8(readAddr >> 8),
			uint8(readAddr >> 16),
			uint8(readAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 6,
		Data: [8]uint8{
			uint8(readAddr >> 32),
			uint8(readAddr >> 40),
			uint8(readAddr >> 48),
			uint8(readAddr >> 56),
			uint8(8),
			uint8(0),
			uint8(0),
			uint8(0)}}

	// Transmit the request message.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Accept the response message.
	respFlit1 := <-smiResponse
	respFlit2 := <-smiResponse

	return (((uint64(respFlit1.Data[4])) |
		(uint64(respFlit1.Data[5]) << 8)) |
		((uint64(respFlit1.Data[6]) << 16) |
			(uint64(respFlit1.Data[7]) << 24))) |
		(((uint64(respFlit2.Data[0]) << 32) |
			(uint64(respFlit2.Data[1]) << 40)) |
			((uint64(respFlit2.Data[2]) << 48) |
				(uint64(respFlit2.Data[3]) << 56)))
}

//
// ReadUInt32 reads a single 32-bit unsigned data value from a word aligned
// address on the specified SMI memory endpoint, with the bottom two address
// bits being ignored.
// TODO: The status of the write transaction should also be returned as the
// boolean 'readOk' flag.
//
func ReadUInt32(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddr uintptr,
	readOptions uint8) uint32 {

	// Assemble the request message.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemReadReq),
			uint8(readOptions),
			uint8(0),
			uint8(0),
			uint8(readAddr) & 0xFC,
			uint8(readAddr >> 8),
			uint8(readAddr >> 16),
			uint8(readAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 6,
		Data: [8]uint8{
			uint8(readAddr >> 32),
			uint8(readAddr >> 40),
			uint8(readAddr >> 48),
			uint8(readAddr >> 56),
			uint8(4),
			uint8(0),
			uint8(0),
			uint8(0)}}

	// Transmit the request message.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Accept the response message.
	respFlit1 := <-smiResponse

	return (((uint32(respFlit1.Data[4])) |
		(uint32(respFlit1.Data[5]) << 8)) |
		((uint32(respFlit1.Data[6]) << 16) |
			(uint32(respFlit1.Data[7]) << 24)))
}

//
// ReadUInt16 reads a single 16-bit unsigned data value from a word aligned
// address on the specified SMI memory endpoint, with the bottom address
// bit being ignored.
// TODO: The status of the write transaction should also be returned as the
// boolean 'readOk' flag.
//
func ReadUInt16(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddr uintptr,
	readOptions uint8) uint16 {

	// Assemble the request message.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemReadReq),
			uint8(readOptions),
			uint8(0),
			uint8(0),
			uint8(readAddr) & 0xFE,
			uint8(readAddr >> 8),
			uint8(readAddr >> 16),
			uint8(readAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 6,
		Data: [8]uint8{
			uint8(readAddr >> 32),
			uint8(readAddr >> 40),
			uint8(readAddr >> 48),
			uint8(readAddr >> 56),
			uint8(2),
			uint8(0),
			uint8(0),
			uint8(0)}}

	// Transmit the request message.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Accept the response message.
	respFlit1 := <-smiResponse

	return uint16(respFlit1.Data[4]) |
		(uint16(respFlit1.Data[5]) << 8)
}

//
// ReadUInt8 reads a single 8-bit unsigned data value from a byte aligned
// address on the specified SMI memory endpoint.
// TODO: The status of the write transaction should also be returned as the
// boolean 'readOk' flag.
//
func ReadUInt8(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddr uintptr,
	readOptions uint8) uint8 {

	// Assemble the request message.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemReadReq),
			uint8(readOptions),
			uint8(0),
			uint8(0),
			uint8(readAddr),
			uint8(readAddr >> 8),
			uint8(readAddr >> 16),
			uint8(readAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 6,
		Data: [8]uint8{
			uint8(readAddr >> 32),
			uint8(readAddr >> 40),
			uint8(readAddr >> 48),
			uint8(readAddr >> 56),
			uint8(1),
			uint8(0),
			uint8(0),
			uint8(0)}}

	// Transmit the request message.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Accept the response message.
	respFlit1 := <-smiResponse

	return respFlit1.Data[4]
}

//
// writeSingleBurstUInt64 is the core logic for writing a single incrementing
// burst of 64-bit unsigned data. Requires validated and word aligned input
// parameters.
//
func writeSingleBurstUInt64(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddr uintptr,
	writeOptions uint8,
	writeLength uint16,
	writeDataChan <-chan uint64) bool {

	// Set up the initial flit data.
	firstFlit := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemWriteReq),
			uint8(writeOptions),
			uint8(0),
			uint8(0),
			uint8(writeAddr),
			uint8(writeAddr >> 8),
			uint8(writeAddr >> 16),
			uint8(writeAddr >> 24)}}

	flitData := [6]uint8{
		uint8(writeAddr >> 32),
		uint8(writeAddr >> 40),
		uint8(writeAddr >> 48),
		uint8(writeAddr >> 56),
		uint8(writeLength),
		uint8(writeLength >> 8)}

	// Transmit the initial request flit.
	smiRequest <- firstFlit

	// Pull the requested number of words from the write data channel and
	// write the updated flit data to the request output.
	for i := (writeLength >> 3); i != 0; i-- {
		writeData := <-writeDataChan
		outputFlit := Flit64{
			Eofc: 0,
			Data: [8]uint8{
				flitData[0],
				flitData[1],
				flitData[2],
				flitData[3],
				flitData[4],
				flitData[5],
				uint8(writeData),
				uint8(writeData >> 8)}}
		flitData[0] = uint8(writeData >> 16)
		flitData[1] = uint8(writeData >> 24)
		flitData[2] = uint8(writeData >> 32)
		flitData[3] = uint8(writeData >> 40)
		flitData[4] = uint8(writeData >> 48)
		flitData[5] = uint8(writeData >> 56)
		smiRequest <- outputFlit
	}

	// Send the final flit.
	smiRequest <- Flit64{
		Eofc: 6,
		Data: [8]uint8{
			flitData[0],
			flitData[1],
			flitData[2],
			flitData[3],
			flitData[4],
			flitData[5],
			uint8(0),
			uint8(0)}}

	// Accept the response message.
	respFlit := <-smiResponse
	var writeOk bool
	if (respFlit.Data[1] & 0x02) == uint8(0x00) {
		writeOk = true
	} else {
		writeOk = false
	}
	return writeOk
}

//
// writeSingleBurstUInt32 is the core logic for writing a single incrementing
// burst of 32-bit unsigned data. Requires validated and word aligned input
// parameters.
//
func writeSingleBurstUInt32(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddr uintptr,
	writeOptions uint8,
	writeLength uint16,
	writeDataChan <-chan uint32) bool {

	// Set up the initial flit data.
	firstFlit := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemWriteReq),
			uint8(writeOptions),
			uint8(0),
			uint8(0),
			uint8(writeAddr),
			uint8(writeAddr >> 8),
			uint8(writeAddr >> 16),
			uint8(writeAddr >> 24)}}

	flitData := [6]uint8{
		uint8(writeAddr >> 32),
		uint8(writeAddr >> 40),
		uint8(writeAddr >> 48),
		uint8(writeAddr >> 56),
		uint8(writeLength),
		uint8(writeLength >> 8)}
	finalEofc := uint8(6)

	// Transmit the initial request flit.
	smiRequest <- firstFlit

	// Pull the requested number of words from the write data channel and
	// write the updated flit data to the request output.
	for i := (writeLength >> 2); i != 0; i-- {
		writeData := <-writeDataChan
		if finalEofc == 6 {
			outputFlit := Flit64{
				Eofc: 0,
				Data: [8]uint8{
					flitData[0],
					flitData[1],
					flitData[2],
					flitData[3],
					flitData[4],
					flitData[5],
					uint8(writeData),
					uint8(writeData >> 8)}}
			flitData[0] = uint8(writeData >> 16)
			flitData[1] = uint8(writeData >> 24)
			smiRequest <- outputFlit
			finalEofc = 2
		} else {
			flitData[2] = uint8(writeData)
			flitData[3] = uint8(writeData >> 8)
			flitData[4] = uint8(writeData >> 16)
			flitData[5] = uint8(writeData >> 24)
			finalEofc = 6
		}
	}

	// Send the final flit.
	smiRequest <- Flit64{
		Eofc: finalEofc,
		Data: [8]uint8{
			flitData[0],
			flitData[1],
			flitData[2],
			flitData[3],
			flitData[4],
			flitData[5],
			uint8(0),
			uint8(0)}}

	// Accept the response message.
	respFlit := <-smiResponse
	var writeOk bool
	if (respFlit.Data[1] & 0x02) == uint8(0x00) {
		writeOk = true
	} else {
		writeOk = false
	}
	return writeOk
}

//
// writeSingleBurstUInt16 is the core logic for writing a single incrementing
// burst of 16-bit unsigned data. Requires validated and word aligned input
// parameters.
//
func writeSingleBurstUInt16(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddr uintptr,
	writeOptions uint8,
	writeLength uint16,
	writeDataChan <-chan uint16) bool {

	// Set up the initial flit data.
	firstFlit := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemWriteReq),
			uint8(writeOptions),
			uint8(0),
			uint8(0),
			uint8(writeAddr),
			uint8(writeAddr >> 8),
			uint8(writeAddr >> 16),
			uint8(writeAddr >> 24)}}

	flitData := [8]uint8{
		uint8(writeAddr >> 32),
		uint8(writeAddr >> 40),
		uint8(writeAddr >> 48),
		uint8(writeAddr >> 56),
		uint8(writeLength),
		uint8(writeLength >> 8),
		uint8(0),
		uint8(0)}
	finalEofc := uint8(6)

	// Transmit the initial request flit.
	smiRequest <- firstFlit

	// Pull the requested number of words from the write data channel and
	// write the updated flit data to the request output.
	for i := (writeLength >> 1); i != 0; i-- {
		writeData := <-writeDataChan
		switch finalEofc {
		case 2:
			flitData[2] = uint8(writeData)
			flitData[3] = uint8(writeData >> 8)
			finalEofc = 4
		case 4:
			flitData[4] = uint8(writeData)
			flitData[5] = uint8(writeData >> 8)
			finalEofc = 6
		case 6:
			flitData[6] = uint8(writeData)
			flitData[7] = uint8(writeData >> 8)
			finalEofc = 8
		default:
			outputFlit := Flit64{
				Eofc: 0,
				Data: flitData}
			flitData[0] = uint8(writeData)
			flitData[1] = uint8(writeData >> 8)
			smiRequest <- outputFlit
			finalEofc = 2
		}
	}

	// Send the final flit.
	smiRequest <- Flit64{
		Eofc: finalEofc,
		Data: flitData}

	// Accept the response message.
	respFlit := <-smiResponse
	var writeOk bool
	if (respFlit.Data[1] & 0x02) == uint8(0x00) {
		writeOk = true
	} else {
		writeOk = false
	}
	return writeOk
}

//
// writeSingleBurstUInt8 is the core logic for writing a single incrementing
// burst of 8-bit unsigned data. Requires validated input parameters.
//
func writeSingleBurstUInt8(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddr uintptr,
	writeOptions uint8,
	writeLength uint16,
	writeDataChan <-chan uint8) bool {

	// Set up the initial flit data.
	firstFlit := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemWriteReq),
			uint8(writeOptions),
			uint8(0),
			uint8(0),
			uint8(writeAddr),
			uint8(writeAddr >> 8),
			uint8(writeAddr >> 16),
			uint8(writeAddr >> 24)}}

	flitData := [8]uint8{
		uint8(writeAddr >> 32),
		uint8(writeAddr >> 40),
		uint8(writeAddr >> 48),
		uint8(writeAddr >> 56),
		uint8(writeLength),
		uint8(writeLength >> 8),
		uint8(0),
		uint8(0)}
	finalEofc := uint8(6)

	// Transmit the initial request flit.
	smiRequest <- firstFlit

	// Pull the requested number of words from the write data channel and
	// write the updated flit data to the request output.
	for i := (writeLength); i != 0; i-- {
		writeData := <-writeDataChan
		switch finalEofc {
		case 1:
			flitData[1] = writeData
			finalEofc = 2
		case 2:
			flitData[2] = writeData
			finalEofc = 3
		case 3:
			flitData[3] = writeData
			finalEofc = 4
		case 4:
			flitData[4] = writeData
			finalEofc = 5
		case 5:
			flitData[5] = writeData
			finalEofc = 6
		case 6:
			flitData[6] = writeData
			finalEofc = 7
		case 7:
			flitData[7] = writeData
			finalEofc = 8
		default:
			outputFlit := Flit64{
				Eofc: 0,
				Data: flitData}
			flitData[0] = writeData
			smiRequest <- outputFlit
			finalEofc = 1
		}
	}

	// Send the final flit.
	smiRequest <- Flit64{
		Eofc: finalEofc,
		Data: flitData}

	// Accept the response message.
	respFlit := <-smiResponse
	var writeOk bool
	if (respFlit.Data[1] & 0x02) == uint8(0x00) {
		writeOk = true
	} else {
		writeOk = false
	}
	return writeOk
}

//
// WritePagedBurstUInt64 writes an incrementing burst of 64-bit unsigned data
// values to a word aligned address on the specified SMI memory endpoint, with
// the bottom three address bits being ignored. The supplied burst length
// specifies the number of 64-bit values to be transferred. The overall burst
// must be contained within a single 4096 byte page and must not cross page
// boundaries. In order to ensure optimum performance, the write data channel
// should be a buffered channel that already contains all the data to be
// written prior to invoking this function. The status of the write transaction
// is returned as the boolean 'writeOk' flag.
//
func WritePagedBurstUInt64(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddrIn uintptr,
	writeOptions uint8,
	writeLengthIn uint16,
	writeDataChan <-chan uint64) bool {

	// TODO: Page boundary validation.
	// Force word alignment.
	writeAddr := writeAddrIn & 0xFFFFFFFFFFFFFFF8
	writeLength := writeLengthIn << 3

	return writeSingleBurstUInt64(
		smiRequest, smiResponse, writeAddr, writeOptions, writeLength, writeDataChan)
}

//
// WritePagedBurstUInt32 writes an incrementing burst of 32-bit unsigned data
// values to a word aligned address on the specified SMI memory endpoint, with
// the bottom two address bits being ignored. The supplied burst length
// specifies the number of 32-bit values to be transferred. The overall burst
// must be contained within a single 4096 byte page and must not cross page
// boundaries. In order to ensure optimum performance, the write data channel
// should be a buffered channel that already contains all the data to be
// written prior to invoking this function. The status of the write transaction
// is returned as the boolean 'writeOk' flag.
//
func WritePagedBurstUInt32(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddrIn uintptr,
	writeOptions uint8,
	writeLengthIn uint16,
	writeDataChan <-chan uint32) bool {

	// TODO: Page boundary validation.
	// Force word alignment.
	writeAddr := writeAddrIn & 0xFFFFFFFFFFFFFFFC
	writeLength := writeLengthIn << 2

	return writeSingleBurstUInt32(
		smiRequest, smiResponse, writeAddr, writeOptions, writeLength, writeDataChan)
}

//
// WritePagedBurstUInt16 writes an incrementing burst of 16-bit unsigned data
// values to a word aligned address on the specified SMI memory endpoint, with
// the bottom address bit being ignored. The supplied burst length specifies
// the number of 16-bit values to be transferred. The overall burst must be
// contained within a single 4096 byte page and must not cross page boundaries.
// In order to ensure optimum performance, the write data channel should be a
// buffered channel that already contains all the data to be written prior to
// invoking this function. The status of the write transaction is returned as
// the boolean 'writeOk' flag.
//
func WritePagedBurstUInt16(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddrIn uintptr,
	writeOptions uint8,
	writeLengthIn uint16,
	writeDataChan <-chan uint16) bool {

	// TODO: Page boundary validation.
	// Force word alignment.
	writeAddr := writeAddrIn & 0xFFFFFFFFFFFFFFFE
	writeLength := writeLengthIn << 1

	return writeSingleBurstUInt16(
		smiRequest, smiResponse, writeAddr, writeOptions, writeLength, writeDataChan)
}

//
// WritePagedBurstUInt8 writes an incrementing burst of 8-bit unsigned data
// values to a byte aligned address on the specified SMI memory endpoint. The
// burst must be contained within a single 4096 byte page and must not cross
// page boundaries. In order to ensure optimum performance, the write data
// channel should be a buffered channel that already contains all the data to
// be written prior to invoking this function. The status of the write
// transaction is returned as the boolean 'writeOk' flag.
//
func WritePagedBurstUInt8(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddrIn uintptr,
	writeOptions uint8,
	writeLengthIn uint16,
	writeDataChan <-chan uint8) bool {

	// TODO: Page boundary validation.

	return writeSingleBurstUInt8(
		smiRequest, smiResponse, writeAddrIn, writeOptions, writeLengthIn, writeDataChan)
}

//
// WriteBurstUInt64 writes an incrementing burst of 64-bit unsigned data
// values to a word aligned address on the specified SMI memory endpoint, with
// the bottom three address bits being ignored. The supplied burst length
// specifies the number of 64-bit values to be transferred, up to a maximum of
// 2^29-1. The burst is automatically segmented to respect page boundaries and
// avoid blocking other transactions. In order to ensure optimum performance,
// the write data channel should be a buffered channel that already contains
// all the data to be written prior to invoking this function. The status of
// the write transaction is returned as the boolean 'writeOk' flag.
//
func WriteBurstUInt64(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddrIn uintptr,
	writeOptions uint8,
	writeLengthIn uint32,
	writeDataChan <-chan uint64) bool {

	writeOk := true
	writeAddr := writeAddrIn & 0xFFFFFFFFFFFFFFF8
	writeLength := writeLengthIn << 3
	burstOffset := uint16(writeAddr) & uint16(SmiMemBurstSize-1)
	burstSize := uint16(SmiMemBurstSize) - burstOffset
	smiWriteChan := make(chan Flit64, 1)

	for writeLength != 0 {
		go AssembleFrame64(smiWriteChan, smiRequest)
		if writeLength < uint32(burstSize) {
			burstSize = uint16(writeLength)
		}
		writeOk = writeOk && writeSingleBurstUInt64(
			smiWriteChan, smiResponse, writeAddr, writeOptions, burstSize, writeDataChan)
		writeAddr += uintptr(burstSize)
		writeLength -= uint32(burstSize)
		burstSize = uint16(SmiMemBurstSize)
	}
	return writeOk
}

//
// WriteBurstUInt32 writes an incrementing burst of 32-bit unsigned data
// values to a word aligned address on the specified SMI memory endpoint, with
// the bottom two address bits being ignored. The supplied burst length
// specifies the number of 32-bit values to be transferred, up to a maximum of
// 2^30-1. The burst is automatically segmented to respect page boundaries and
// avoid blocking other transactions. In order to ensure optimum performance,
// the write data channel should be a buffered channel that already contains
// all the data to be written prior to invoking this function. The status of
// the write transaction is returned as the boolean 'writeOk' flag.
//
func WriteBurstUInt32(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddrIn uintptr,
	writeOptions uint8,
	writeLengthIn uint32,
	writeDataChan <-chan uint32) bool {

	writeOk := true
	writeAddr := writeAddrIn & 0xFFFFFFFFFFFFFFFC
	writeLength := writeLengthIn << 2
	burstOffset := uint16(writeAddr) & uint16(SmiMemBurstSize-1)
	burstSize := uint16(SmiMemBurstSize) - burstOffset
	smiWriteChan := make(chan Flit64, 1)

	for writeLength != 0 {
		go AssembleFrame64(smiWriteChan, smiRequest)
		if writeLength < uint32(burstSize) {
			burstSize = uint16(writeLength)
		}
		writeOk = writeOk && writeSingleBurstUInt32(
			smiWriteChan, smiResponse, writeAddr, writeOptions, burstSize, writeDataChan)
		writeAddr += uintptr(burstSize)
		writeLength -= uint32(burstSize)
		burstSize = uint16(SmiMemBurstSize)
	}
	return writeOk
}

//
// WriteBurstUInt16 writes an incrementing burst of 16-bit unsigned data
// values to a word aligned address on the specified SMI memory endpoint, with
// the bottom address bit being ignored. The supplied burst length specifies
// the number of 16-bit values to be transferred, up to a maximum of 2^31-1.
// The burst is automatically segmented to respect page boundaries and avoid
// blocking other transactions. In order to ensure optimum performance, the
// write data channel should be a buffered channel that already contains all
// the data to be written prior to invoking this function. The status of the
// write transaction is returned as the boolean 'writeOk' flag.
//
func WriteBurstUInt16(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddrIn uintptr,
	writeOptions uint8,
	writeLengthIn uint32,
	writeDataChan <-chan uint16) bool {

	writeOk := true
	writeAddr := writeAddrIn & 0xFFFFFFFFFFFFFFFE
	writeLength := writeLengthIn << 1
	burstOffset := uint16(writeAddr) & uint16(SmiMemBurstSize-1)
	burstSize := uint16(SmiMemBurstSize) - burstOffset
	smiWriteChan := make(chan Flit64, 1)

	for writeLength != 0 {
		go AssembleFrame64(smiWriteChan, smiRequest)
		if writeLength < uint32(burstSize) {
			burstSize = uint16(writeLength)
		}
		writeOk = writeOk && writeSingleBurstUInt16(
			smiWriteChan, smiResponse, writeAddr, writeOptions, burstSize, writeDataChan)
		writeAddr += uintptr(burstSize)
		writeLength -= uint32(burstSize)
		burstSize = uint16(SmiMemBurstSize)
	}
	return writeOk
}

//
// WriteBurstUInt8 writes an incrementing burst of 8-bit unsigned data
// values to a byte aligned address on the specified SMI memory endpoint. The
// burst is automatically segmented to respect page boundaries and avoid
// blocking other transactions. In order to ensure optimum performance, the
// write data channel should be a buffered channel that already contains all
// the data to be written prior to invoking this function. The status of the
// write transaction is returned as the boolean 'writeOk' flag.
//
func WriteBurstUInt8(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	writeAddrIn uintptr,
	writeOptions uint8,
	writeLengthIn uint32,
	writeDataChan <-chan uint8) bool {

	writeOk := true
	writeAddr := writeAddrIn
	writeLength := writeLengthIn
	burstOffset := uint16(writeAddr) & uint16(SmiMemBurstSize-1)
	burstSize := uint16(SmiMemBurstSize) - burstOffset
	smiWriteChan := make(chan Flit64, 1)

	for writeLength != 0 {
		go AssembleFrame64(smiWriteChan, smiRequest)
		if writeLength < uint32(burstSize) {
			burstSize = uint16(writeLength)
		}
		writeOk = writeOk && writeSingleBurstUInt8(
			smiWriteChan, smiResponse, writeAddr, writeOptions, burstSize, writeDataChan)
		writeAddr += uintptr(burstSize)
		writeLength -= uint32(burstSize)
		burstSize = uint16(SmiMemBurstSize)
	}
	return writeOk
}

//
// readSingleBurstUInt64 is the core logic for reading a single incrementing
// burst of 64-bit unsigned data. Requires validated and word aligned input
// parameters.
//
func readSingleBurstUInt64(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddr uintptr,
	readOptions uint8,
	readLength uint16,
	readDataChan chan<- uint64) bool {

	// Set up the request flit data.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemReadReq),
			uint8(readOptions),
			uint8(0),
			uint8(0),
			uint8(readAddr),
			uint8(readAddr >> 8),
			uint8(readAddr >> 16),
			uint8(readAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 6,
		Data: [8]uint8{
			uint8(readAddr >> 32),
			uint8(readAddr >> 40),
			uint8(readAddr >> 48),
			uint8(readAddr >> 56),
			uint8(readLength),
			uint8(readLength >> 8),
			uint8(0),
			uint8(0)}}

	// Transmit the request flits.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Pull the response header flit from the response channel
	respFlit1 := <-smiResponse
	flitData := [4]uint8{
		respFlit1.Data[4],
		respFlit1.Data[5],
		respFlit1.Data[6],
		respFlit1.Data[7]}
	moreFlits := respFlit1.Eofc == 0

	var readOk bool
	if (respFlit1.Data[1] & 0x02) == uint8(0x00) {
		readOk = true
	} else {
		readOk = false
	}

	// Pull all the payload flits from the response channel and copy the data
	// to the output channel.
	for moreFlits {
		respFlitN := <-smiResponse
		readDataVal :=
			((uint64(flitData[0]) |
				(uint64(flitData[1]) << 8)) |
				((uint64(flitData[2]) << 16) |
					(uint64(flitData[3]) << 24))) |
				(((uint64(respFlitN.Data[0]) << 32) |
					(uint64(respFlitN.Data[1]) << 40)) |
					((uint64(respFlitN.Data[2]) << 48) |
						(uint64(respFlitN.Data[3]) << 56)))
		flitData = [4]uint8{
			respFlitN.Data[4],
			respFlitN.Data[5],
			respFlitN.Data[6],
			respFlitN.Data[7]}
		moreFlits = respFlitN.Eofc == 0
		readDataChan <- readDataVal
	}
	return readOk
}

//
// readSingleBurstUInt32 is the core logic for reading a single incrementing
// burst of 32-bit unsigned data. Requires validated and word aligned input
// parameters.
//
func readSingleBurstUInt32(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddr uintptr,
	readOptions uint8,
	readLength uint16,
	readDataChan chan<- uint32) bool {

	// Set up the request flit data.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemReadReq),
			uint8(readOptions),
			uint8(0),
			uint8(0),
			uint8(readAddr),
			uint8(readAddr >> 8),
			uint8(readAddr >> 16),
			uint8(readAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 6,
		Data: [8]uint8{
			uint8(readAddr >> 32),
			uint8(readAddr >> 40),
			uint8(readAddr >> 48),
			uint8(readAddr >> 56),
			uint8(readLength),
			uint8(readLength >> 8),
			uint8(0),
			uint8(0)}}

	// Transmit the request flits.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Pull the response header flit from the response channel
	respFlit1 := <-smiResponse
	flitData := [4]uint8{
		respFlit1.Data[4],
		respFlit1.Data[5],
		respFlit1.Data[6],
		respFlit1.Data[7]}
	readOffset := uint8(4)

	var readOk bool
	if (respFlit1.Data[1] & 0x02) == uint8(0x00) {
		readOk = true
	} else {
		readOk = false
	}

	// Pull all the payload flits from the response channel and copy the data
	// to the output channel.
	for i := (readLength >> 2); i != 0; i-- {
		var readData uint32
		if readOffset == 4 {
			readData =
				(uint32(flitData[0]) |
					(uint32(flitData[1]) << 8)) |
					((uint32(flitData[2]) << 16) |
						(uint32(flitData[3]) << 24))
			readOffset = 0
		} else {
			respFlitN := <-smiResponse
			flitData = [4]uint8{
				respFlitN.Data[4],
				respFlitN.Data[5],
				respFlitN.Data[6],
				respFlitN.Data[7]}
			readData =
				(uint32(respFlitN.Data[0]) |
					(uint32(respFlitN.Data[1]) << 8)) |
					((uint32(respFlitN.Data[2]) << 16) |
						(uint32(respFlitN.Data[3]) << 24))
			readOffset = 4
		}
		readDataChan <- readData
	}
	return readOk
}

//
// readSingleBurstUInt16 is the core logic for reading a single incrementing
// burst of 16-bit unsigned data. Requires validated and word aligned input
// parameters.
//
func readSingleBurstUInt16(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddr uintptr,
	readOptions uint8,
	readLength uint16,
	readDataChan chan<- uint16) bool {

	// Set up the request flit data.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemReadReq),
			uint8(readOptions),
			uint8(0),
			uint8(0),
			uint8(readAddr),
			uint8(readAddr >> 8),
			uint8(readAddr >> 16),
			uint8(readAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 6,
		Data: [8]uint8{
			uint8(readAddr >> 32),
			uint8(readAddr >> 40),
			uint8(readAddr >> 48),
			uint8(readAddr >> 56),
			uint8(readLength),
			uint8(readLength >> 8),
			uint8(0),
			uint8(0)}}

	// Transmit the request flits.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Pull the response header flit from the response channel
	respFlit1 := <-smiResponse
	flitData := [6]uint8{
		uint8(0),
		uint8(0),
		respFlit1.Data[4],
		respFlit1.Data[5],
		respFlit1.Data[6],
		respFlit1.Data[7]}
	readOffset := uint8(4)

	var readOk bool
	if (respFlit1.Data[1] & 0x02) == uint8(0x00) {
		readOk = true
	} else {
		readOk = false
	}

	// Pull all the payload flits from the response channel and copy the data
	// to the output channel.
	for i := (readLength >> 1); i != 0; i-- {
		var readData uint16
		switch readOffset {
		case 2:
			readData =
				(uint16(flitData[0])) |
					(uint16(flitData[1]) << 8)
			readOffset = 4
		case 4:
			readData =
				(uint16(flitData[2])) |
					(uint16(flitData[3]) << 8)
			readOffset = 6
		case 6:
			readData =
				(uint16(flitData[4])) |
					(uint16(flitData[5]) << 8)
			readOffset = 0
		default:
			respFlitN := <-smiResponse
			flitData = [6]uint8{
				respFlitN.Data[2],
				respFlitN.Data[3],
				respFlitN.Data[4],
				respFlitN.Data[5],
				respFlitN.Data[6],
				respFlitN.Data[7]}
			readData =
				(uint16(respFlitN.Data[0])) |
					(uint16(respFlitN.Data[1]) << 8)
			readOffset = 2
		}
		readDataChan <- readData
	}
	return readOk
}

//
// readSingleBurstUInt8 is the core logic for reading a single incrementing
// burst of 8-bit unsigned data. Requires validated input parameters.
//
func readSingleBurstUInt8(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddr uintptr,
	readOptions uint8,
	readLength uint16,
	readDataChan chan<- uint8) bool {

	// Set up the request flit data.
	reqFlit1 := Flit64{
		Eofc: 0,
		Data: [8]uint8{
			uint8(SmiMemReadReq),
			uint8(readOptions),
			uint8(0),
			uint8(0),
			uint8(readAddr),
			uint8(readAddr >> 8),
			uint8(readAddr >> 16),
			uint8(readAddr >> 24)}}

	reqFlit2 := Flit64{
		Eofc: 6,
		Data: [8]uint8{
			uint8(readAddr >> 32),
			uint8(readAddr >> 40),
			uint8(readAddr >> 48),
			uint8(readAddr >> 56),
			uint8(readLength),
			uint8(readLength >> 8),
			uint8(0),
			uint8(0)}}

	// Transmit the request flits.
	smiRequest <- reqFlit1
	smiRequest <- reqFlit2

	// Pull the response header flit from the response channel
	respFlit1 := <-smiResponse
	flitData := [7]uint8{
		uint8(0),
		uint8(0),
		uint8(0),
		respFlit1.Data[4],
		respFlit1.Data[5],
		respFlit1.Data[6],
		respFlit1.Data[7]}
	readOffset := uint8(4)

	var readOk bool
	if (respFlit1.Data[1] & 0x02) == uint8(0x00) {
		readOk = true
	} else {
		readOk = false
	}

	// Pull all the payload flits from the response channel and copy the data
	// to the output channel.
	for i := readLength; i != 0; i-- {
		var readData uint8
		switch readOffset {
		case 1:
			readData = flitData[0]
			readOffset = 2
		case 2:
			readData = flitData[1]
			readOffset = 3
		case 3:
			readData = flitData[2]
			readOffset = 4
		case 4:
			readData = flitData[3]
			readOffset = 5
		case 5:
			readData = flitData[4]
			readOffset = 6
		case 6:
			readData = flitData[5]
			readOffset = 7
		case 7:
			readData = flitData[6]
			readOffset = 0
		default:
			respFlitN := <-smiResponse
			flitData = [7]uint8{
				respFlitN.Data[1],
				respFlitN.Data[2],
				respFlitN.Data[3],
				respFlitN.Data[4],
				respFlitN.Data[5],
				respFlitN.Data[6],
				respFlitN.Data[7]}
			readData =
				respFlitN.Data[0]
			readOffset = 1
		}
		readDataChan <- readData
	}
	return readOk
}

//
// ReadPagedBurstUInt64 reads an incrementing burst of 64-bit unsigned data
// values from a word aligned address on the specified SMI memory endpoint,
// with the bottom three address bits being ignored. The supplied burst length
// specifies the number of 64-bit values to be transferred. The overall burst
// must be contained within a single 4096 byte page and must not cross page
// boundaries. In order to ensure optimum performance, the read data channel
// should be a buffered channel that has sufficient free space to hold all the
// data to be transferred. The status of the read transaction is returned as
// the boolean 'readOk' flag.
//
func ReadPagedBurstUInt64(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddrIn uintptr,
	readOptions uint8,
	readLengthIn uint16,
	readDataChan chan<- uint64) bool {

	// TODO: Page boundary validation.
	// Force word alignment.
	readAddr := readAddrIn & 0xFFFFFFFFFFFFFFF8
	readLength := readLengthIn << 3

	return readSingleBurstUInt64(
		smiRequest, smiResponse, readAddr, readOptions, readLength, readDataChan)
}

//
// ReadPagedBurstUInt32 reads an incrementing burst of 32-bit unsigned data
// values from a word aligned address on the specified SMI memory endpoint,
// with the bottom two address bits being ignored. The supplied burst length
// specifies the number of 32-bit values to be transferred. The overall burst
// must be contained within a single 4096 byte page and must not cross page
// boundaries. In order to ensure optimum performance, the read data channel
// should be a buffered channel that has sufficient free space to hold all the
// data to be transferred. The status of the read transaction is returned as
// the boolean 'readOk' flag.
//
func ReadPagedBurstUInt32(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddrIn uintptr,
	readOptions uint8,
	readLengthIn uint16,
	readDataChan chan<- uint32) bool {

	// TODO: Page boundary validation.
	// Force word alignment.
	readAddr := readAddrIn & 0xFFFFFFFFFFFFFFFC
	readLength := readLengthIn << 2

	return readSingleBurstUInt32(
		smiRequest, smiResponse, readAddr, readOptions, readLength, readDataChan)
}

//
// ReadPagedBurstUInt16 reads an incrementing burst of 16-bit unsigned data
// values from a word aligned address on the specified SMI memory endpoint,
// with the bottom address bit being ignored. The supplied burst length
// specifies the number of 16-bit values to be transferred. The overall burst
// must be contained within a single 4096 byte page and must not cross page
// boundaries. In order to ensure optimum performance, the read data channel
// should be a buffered channel that has sufficient free space to hold all the
// data to be transferred. The status of the read transaction is returned as
// the boolean 'readOk' flag.
//
func ReadPagedBurstUInt16(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddrIn uintptr,
	readOptions uint8,
	readLengthIn uint16,
	readDataChan chan<- uint16) bool {

	// TODO: Page boundary validation.
	// Force word alignment.
	readAddr := readAddrIn & 0xFFFFFFFFFFFFFFFE
	readLength := readLengthIn << 1

	return readSingleBurstUInt16(
		smiRequest, smiResponse, readAddr, readOptions, readLength, readDataChan)
}

//
// ReadPagedBurstUInt8 reads an incrementing burst of 8-bit unsigned data
// values from a byte aligned address on the specified SMI memory endpoint.
// The burst must be contained within a single 4096 byte page and must not
// cross page boundaries. In order to ensure optimum performance, the read
// data channel should be a buffered channel that has sufficient free space to
// hold all the data to be transferred. The status of the read transaction is
// returned as the boolean 'readOk' flag.
//
func ReadPagedBurstUInt8(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddrIn uintptr,
	readOptions uint8,
	readLengthIn uint16,
	readDataChan chan<- uint8) bool {

	// TODO: Page boundary validation.

	return readSingleBurstUInt8(
		smiRequest, smiResponse, readAddrIn, readOptions, readLengthIn, readDataChan)
}

//
// ReadBurstUInt64 reads an incrementing burst of 64-bit unsigned data
// values from a word aligned address on the specified SMI memory endpoint,
// with the bottom three address bits being ignored. The supplied burst length
// specifies the number of 64-bit values to be transferred, up to a maximum of
// 2^29-1. The burst is automatically segmented to respect page boundaries and
// avoid blocking other transactions. In order to ensure optimum performance,
// the read data channel should be a buffered channel that has sufficient free
// space to hold all the data to be transferred. The status of the read
// transaction is returned as the boolean 'readOk' flag.
//
func ReadBurstUInt64(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddrIn uintptr,
	readOptions uint8,
	readLengthIn uint32,
	readDataChan chan<- uint64) bool {

	readOk := true
	readAddr := readAddrIn & 0xFFFFFFFFFFFFFFF8
	readLength := readLengthIn << 3
	burstOffset := uint16(readAddr) & uint16(SmiMemBurstSize-1)
	burstSize := uint16(SmiMemBurstSize) - burstOffset
	smiReadChan := make(chan Flit64, 1)

	for readLength != 0 {
		go ForwardFrame64(smiResponse, smiReadChan)
		if readLength < uint32(burstSize) {
			burstSize = uint16(readLength)
		}
		readOk = readOk && readSingleBurstUInt64(
			smiRequest, smiReadChan, readAddr, readOptions, burstSize, readDataChan)
		readAddr += uintptr(burstSize)
		readLength -= uint32(burstSize)
		burstSize = uint16(SmiMemBurstSize)
	}
	return readOk
}

//
// ReadBurstUInt32 reads an incrementing burst of 32-bit unsigned data
// values from a word aligned address on the specified SMI memory endpoint,
// with the bottom two address bits being ignored. The supplied burst length
// specifies the number of 32-bit values to be transferred, up to a maximum of
// 2^30-1. The burst is automatically segmented to respect page boundaries and
// avoid blocking other transactions. In order to ensure optimum performance,
// the read data channel should be a buffered channel that has sufficient free
// space to hold all the data to be transferred. The status of the read
// transaction is returned as the boolean 'readOk' flag.
//
func ReadBurstUInt32(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddrIn uintptr,
	readOptions uint8,
	readLengthIn uint32,
	readDataChan chan<- uint32) bool {

	readOk := true
	readAddr := readAddrIn & 0xFFFFFFFFFFFFFFFC
	readLength := readLengthIn << 2
	burstOffset := uint16(readAddr) & uint16(SmiMemBurstSize-1)
	burstSize := uint16(SmiMemBurstSize) - burstOffset
	smiReadChan := make(chan Flit64, 1)

	for readLength != 0 {
		go ForwardFrame64(smiResponse, smiReadChan)
		if readLength < uint32(burstSize) {
			burstSize = uint16(readLength)
		}
		readOk = readOk && readSingleBurstUInt32(
			smiRequest, smiReadChan, readAddr, readOptions, burstSize, readDataChan)
		readAddr += uintptr(burstSize)
		readLength -= uint32(burstSize)
		burstSize = uint16(SmiMemBurstSize)
	}
	return readOk
}

//
// ReadBurstUInt16 reads an incrementing burst of 16-bit unsigned data
// values from a word aligned address on the specified SMI memory endpoint,
// with the bottom address bit being ignored. The supplied burst length
// specifies the number of 16-bit values to be transferred, up to a maximum of
// 2^31-1. The burst is automatically segmented to respect page boundaries and
// avoid blocking other transactions. In order to ensure optimum performance,
// the read data channel should be a buffered channel that has sufficient free
// space to hold all the data to be transferred. The status of the read
// transaction is returned as the boolean 'readOk' flag.
//
func ReadBurstUInt16(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddrIn uintptr,
	readOptions uint8,
	readLengthIn uint32,
	readDataChan chan<- uint16) bool {

	readOk := true
	readAddr := readAddrIn & 0xFFFFFFFFFFFFFFFE
	readLength := readLengthIn << 1
	burstOffset := uint16(readAddr) & uint16(SmiMemBurstSize-1)
	burstSize := uint16(SmiMemBurstSize) - burstOffset
	smiReadChan := make(chan Flit64, 1)

	for readLength != 0 {
		go ForwardFrame64(smiResponse, smiReadChan)
		if readLength < uint32(burstSize) {
			burstSize = uint16(readLength)
		}
		readOk = readOk && readSingleBurstUInt16(
			smiRequest, smiReadChan, readAddr, readOptions, burstSize, readDataChan)
		readAddr += uintptr(burstSize)
		readLength -= uint32(burstSize)
		burstSize = uint16(SmiMemBurstSize)
	}
	return readOk
}

//
// ReadBurstUInt8 reads an incrementing burst of 8-bit unsigned data values
// from a byte aligned address on the specified SMI memory endpoint. The burst
// is automatically segmented to respect page boundaries and avoid blocking
// other transactions. In order to ensure optimum performance, the read data
// channel should be a buffered channel that has sufficient free space to
// hold all the data to be transferred. The status of the read transaction
// is returned as the boolean 'readOk' flag.
//
func ReadBurstUInt8(
	smiRequest chan<- Flit64,
	smiResponse <-chan Flit64,
	readAddrIn uintptr,
	readOptions uint8,
	readLengthIn uint32,
	readDataChan chan<- uint8) bool {

	readOk := true
	readAddr := readAddrIn
	readLength := readLengthIn
	burstOffset := uint16(readAddr) & uint16(SmiMemBurstSize-1)
	burstSize := uint16(SmiMemBurstSize) - burstOffset
	smiReadChan := make(chan Flit64, 1)

	for readLength != 0 {
		go ForwardFrame64(smiResponse, smiReadChan)
		if readLength < uint32(burstSize) {
			burstSize = uint16(readLength)
		}
		readOk = readOk && readSingleBurstUInt8(
			smiRequest, smiReadChan, readAddr, readOptions, burstSize, readDataChan)
		readAddr += uintptr(burstSize)
		readLength -= uint32(burstSize)
		burstSize = uint16(SmiMemBurstSize)
	}
	return readOk
}
