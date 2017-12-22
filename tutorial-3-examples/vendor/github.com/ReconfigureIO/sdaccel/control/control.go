//
// (c) 2017 ReconfigureIO
//
// <COPYRIGHT TERMS>
//

//
// AXI-Lite interface definitions for interactive kernel control transactions.
//

package control

// Specifies AXI-Lite address channel fields.
type Addr struct {
	Addr  uint32
	Cache [4]bool
	Prot  [3]bool
}

// Specifies AXI-Lite read data channel fields.
type ReadData struct {
	Data uint32
	Resp [2]bool
}

// Specifies AXI-Lite write data channel fields.
type WriteData struct {
	Data uint32
	Strb [4]bool
}

// Specifies AXI-Lite write response channel fields.
type WriteResp struct {
	Resp [2]bool
}

// Goroutine to disable control bus read transactions. Should only be run
// once for each control interface.
func DisableReads(controlReadAddr <-chan Addr,
	controlReadData chan<- ReadData) {
	for {
		<-controlReadAddr
		controlReadData <- ReadData{}
	}
}

// Goroutine to disable control bus write transactions. Should only be run once
// for each control interface.
func DisableWrites(
	controlWriteAddr <-chan Addr,
	controlWriteData <-chan WriteData,
	controlWriteResp chan<- WriteResp) {

	for {
		<-controlWriteAddr
		<-controlWriteData
		controlWriteResp <- WriteResp{}
	}
}

// Goroutine to disable control bus parameter RAM accesses. Should only be run
// once for each control interface.
func DisableParams(
	paramAddr chan<- uint32,
	paramData <-chan uint32) {
	paramAddr <- 0
	for {
		<-paramData
	}
}
