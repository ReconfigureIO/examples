//
// (c) 2017 ReconfigureIO
//
// <COPYRIGHT TERMS>
//

//
// AXI protocol bus arbitration between multiple 'upstream' ports. This package
// specifies a set of goroutines which may be used to arbitrate between multiple
// upstream AXI 'server' ports and a single downstream 'client' port. The
// current implementation supports arbitration between 2, 3 or 4 upstream ports.
// TODO: Support arbitrary number of upstream ports on demand using the Go
// generate capability.
//

/*
Package arbitrate provides reusable arbitrators for AXI transations.
*/
package arbitrate

import (
	"github.com/ReconfigureIO/sdaccel/axi/protocol"
)

//
// Goroutine which implements AXI arbitration between two AXI write interfaces.
//
func WriteArbitrateX2(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	serverAddr0 <-chan protocol.Addr,
	serverData0 <-chan protocol.WriteData,
	serverResp0 chan<- protocol.WriteResp,
	serverAddr1 <-chan protocol.Addr,
	serverData1 <-chan protocol.WriteData,
	serverResp1 chan<- protocol.WriteResp) {

	// Specify the input selection channels.
	dataChanSelect := make(chan byte)
	respChanSelect := make(chan byte)

	// Run write data channel handler.
	go func() {
		for {
			var writeData protocol.WriteData
			chanSelect := <-dataChanSelect

			// Terminate transfers on write data channel 'last' flag.
			isLast := false
			for !isLast {
				switch chanSelect {
				case 0:
					writeData = <-serverData0
				default:
					writeData = <-serverData1
				}
				clientData <- writeData
				isLast = writeData.Last
			}
		}
	}()

	// Run response channel handler.
	go func() {
		for {
			chanSelect := <-respChanSelect
			writeResp := <-clientResp
			switch chanSelect {
			case 0:
				serverResp0 <- writeResp
			default:
				serverResp1 <- writeResp
			}
		}
	}()

	// Use intermediate variables for efficient implementation.
	var writeAddr protocol.Addr
	var dataChanId byte
	for {
		select {
		case writeAddr = <-serverAddr0:
			dataChanId = 0
		case writeAddr = <-serverAddr1:
			dataChanId = 1
		}
		clientAddr <- writeAddr
		dataChanSelect <- dataChanId
		respChanSelect <- dataChanId
	}
}

//
// Goroutine which implements AXI arbitration between three AXI write interfaces.
//
func WriteArbitrateX3(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	serverAddr0 <-chan protocol.Addr,
	serverData0 <-chan protocol.WriteData,
	serverResp0 chan<- protocol.WriteResp,
	serverAddr1 <-chan protocol.Addr,
	serverData1 <-chan protocol.WriteData,
	serverResp1 chan<- protocol.WriteResp,
	serverAddr2 <-chan protocol.Addr,
	serverData2 <-chan protocol.WriteData,
	serverResp2 chan<- protocol.WriteResp) {

	// Specify the input selection channels.
	dataChanSelect := make(chan byte)
	respChanSelect := make(chan byte)

	// Run write data channel handler.
	go func() {
		for {
			var writeData protocol.WriteData
			chanSelect := <-dataChanSelect

			// Terminate transfers on write data channel 'last' flag.
			isLast := false
			for !isLast {
				switch chanSelect {
				case 0:
					writeData = <-serverData0
				case 1:
					writeData = <-serverData1
				default:
					writeData = <-serverData2
				}
				clientData <- writeData
				isLast = writeData.Last
			}
		}
	}()

	// Run response channel handler.
	go func() {
		for {
			chanSelect := <-respChanSelect
			writeResp := <-clientResp
			switch chanSelect {
			case 0:
				serverResp0 <- writeResp
			case 1:
				serverResp1 <- writeResp
			default:
				serverResp2 <- writeResp
			}
		}
	}()

	// Use intermediate variables for efficient implementation.
	var writeAddr protocol.Addr
	var dataChanId byte
	for {
		select {
		case writeAddr = <-serverAddr0:
			dataChanId = 0
		case writeAddr = <-serverAddr1:
			dataChanId = 1
		case writeAddr = <-serverAddr2:
			dataChanId = 2
		}
		clientAddr <- writeAddr
		dataChanSelect <- dataChanId
		respChanSelect <- dataChanId
	}
}

//
// Goroutine which implements AXI arbitration between four AXI write interfaces.
//
func WriteArbitrateX4(
	clientAddr chan<- protocol.Addr,
	clientData chan<- protocol.WriteData,
	clientResp <-chan protocol.WriteResp,
	serverAddr0 <-chan protocol.Addr,
	serverData0 <-chan protocol.WriteData,
	serverResp0 chan<- protocol.WriteResp,
	serverAddr1 <-chan protocol.Addr,
	serverData1 <-chan protocol.WriteData,
	serverResp1 chan<- protocol.WriteResp,
	serverAddr2 <-chan protocol.Addr,
	serverData2 <-chan protocol.WriteData,
	serverResp2 chan<- protocol.WriteResp,
	serverAddr3 <-chan protocol.Addr,
	serverData3 <-chan protocol.WriteData,
	serverResp3 chan<- protocol.WriteResp) {

	// Specify the input selection channels.
	dataChanSelect := make(chan byte)
	respChanSelect := make(chan byte)

	// Run write data channel handler.
	go func() {
		for {
			var writeData protocol.WriteData
			chanSelect := <-dataChanSelect

			// Terminate transfers on write data channel 'last' flag.
			isLast := false
			for !isLast {
				switch chanSelect {
				case 0:
					writeData = <-serverData0
				case 1:
					writeData = <-serverData1
				case 2:
					writeData = <-serverData2
				default:
					writeData = <-serverData3
				}
				clientData <- writeData
				isLast = writeData.Last
			}
		}
	}()

	// Run response channel handler.
	go func() {
		for {
			chanSelect := <-respChanSelect
			writeResp := <-clientResp
			switch chanSelect {
			case 0:
				serverResp0 <- writeResp
			case 1:
				serverResp1 <- writeResp
			case 2:
				serverResp2 <- writeResp
			default:
				serverResp3 <- writeResp
			}
		}
	}()

	// Use intermediate variables for efficient implementation.
	var writeAddr protocol.Addr
	var dataChanId byte
	for {
		select {
		case writeAddr = <-serverAddr0:
			dataChanId = 0
		case writeAddr = <-serverAddr1:
			dataChanId = 1
		case writeAddr = <-serverAddr2:
			dataChanId = 2
		case writeAddr = <-serverAddr3:
			dataChanId = 3
		}
		clientAddr <- writeAddr
		dataChanSelect <- dataChanId
		respChanSelect <- dataChanId
	}
}

//
// Goroutine which implements AXI arbitration between two AXI read interfaces.
//
func ReadArbitrateX2(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	serverAddr0 <-chan protocol.Addr,
	serverData0 chan<- protocol.ReadData,
	serverAddr1 <-chan protocol.Addr,
	serverData1 chan<- protocol.ReadData) {

	// Specify the input selection channel.
	dataChanSelect := make(chan byte)

	// Run read data channel handler.
	go func() {
		for {
			chanSelect := <-dataChanSelect

			// Terminate transfers on write data channel 'last' flag.
			isLast := false
			for !isLast {
				readData := <-clientData
				switch chanSelect {
				case 0:
					serverData0 <- readData
					isLast = readData.Last
				default:
					serverData1 <- readData
					isLast = readData.Last
				}
			}
		}
	}()

	// Use intermediate variables for efficient implementation.
	var readAddr protocol.Addr
	var dataChanId byte
	for {
		select {
		case readAddr = <-serverAddr0:
			dataChanId = 0
		case readAddr = <-serverAddr1:
			dataChanId = 1
		}
		clientAddr <- readAddr
		dataChanSelect <- dataChanId
	}
}

//
// Goroutine which implements AXI arbitration between three AXI read interfaces.
//
func ReadArbitrateX3(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	serverAddr0 <-chan protocol.Addr,
	serverData0 chan<- protocol.ReadData,
	serverAddr1 <-chan protocol.Addr,
	serverData1 chan<- protocol.ReadData,
	serverAddr2 <-chan protocol.Addr,
	serverData2 chan<- protocol.ReadData) {

	// Specify the input selection channel.
	dataChanSelect := make(chan byte)

	// Run read data channel handler.
	go func() {
		for {
			chanSelect := <-dataChanSelect

			// Terminate transfers on write data channel 'last' flag.
			isLast := false
			for !isLast {
				readData := <-clientData
				switch chanSelect {
				case 0:
					serverData0 <- readData
					isLast = readData.Last
				case 1:
					serverData1 <- readData
					isLast = readData.Last
				default:
					serverData2 <- readData
					isLast = readData.Last
				}
			}
		}
	}()

	// Use intermediate variables for efficient implementation.
	var readAddr protocol.Addr
	var dataChanId byte
	for {
		select {
		case readAddr = <-serverAddr0:
			dataChanId = 0
		case readAddr = <-serverAddr1:
			dataChanId = 1
		case readAddr = <-serverAddr2:
			dataChanId = 2
		}
		clientAddr <- readAddr
		dataChanSelect <- dataChanId
	}
}

//
// Goroutine which implements AXI arbitration between four AXI read interfaces.
//
func ReadArbitrateX4(
	clientAddr chan<- protocol.Addr,
	clientData <-chan protocol.ReadData,
	serverAddr0 <-chan protocol.Addr,
	serverData0 chan<- protocol.ReadData,
	serverAddr1 <-chan protocol.Addr,
	serverData1 chan<- protocol.ReadData,
	serverAddr2 <-chan protocol.Addr,
	serverData2 chan<- protocol.ReadData,
	serverAddr3 <-chan protocol.Addr,
	serverData3 chan<- protocol.ReadData) {

	// Specify the input selection channel.
	dataChanSelect := make(chan byte)

	// Run read data channel handler.
	go func() {
		for {
			chanSelect := <-dataChanSelect

			// Terminate transfers on write data channel 'last' flag.
			isLast := false
			for !isLast {
				readData := <-clientData
				switch chanSelect {
				case 0:
					serverData0 <- readData
					isLast = readData.Last
				case 1:
					serverData1 <- readData
					isLast = readData.Last
				case 2:
					serverData2 <- readData
					isLast = readData.Last
				default:
					serverData3 <- readData
					isLast = readData.Last
				}
			}
		}
	}()

	// Use intermediate variables for efficient implementation.
	var readAddr protocol.Addr
	var dataChanId byte
	for {
		select {
		case readAddr = <-serverAddr0:
			dataChanId = 0
		case readAddr = <-serverAddr1:
			dataChanId = 1
		case readAddr = <-serverAddr2:
			dataChanId = 2
		case readAddr = <-serverAddr3:
			dataChanId = 3
		}
		clientAddr <- readAddr
		dataChanSelect <- dataChanId
	}
}
