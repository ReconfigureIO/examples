//
// (c) 2017 ReconfigureIO
//
// <COPYRIGHT TERMS>
//

//
// AXI protocol interface to memory mapped RAM and I/O. This defines the data
// types to be used on the AXI write address (AXI_AW), write data (AXI_W),
// write status response (AXI_B), read address (AXI_RA) and read data (AXI_R)
// channels. The protocol package also includes goroutines for disabling unused
// AXI inferface ports. The data bus width is fixed at 64 bits, which
// corresponds to the largest Go primitive data types.
//

/*

Package protocol provides low level primitives for working the AXI4 protocol

*/
package protocol

//
// Type Addr specifies AXI memory address channel fields.
//
type Addr struct {
	Id     bool
	Addr   uintptr
	Len    byte
	Size   [3]bool
	Burst  [2]bool
	Lock   bool
	Cache  [4]bool
	Prot   [3]bool
	Region [4]bool
	Qos    [4]bool
	User   bool
}

//
// Type ReadData specifies AXI memory read data channel fields.
//
type ReadData struct {
	Id   bool
	Data uint64
	Resp [2]bool
	Last bool
	User bool
}

//
// Type WriteData specifies AXI memory write data channel fields.
//
type WriteData struct {
	Data uint64
	Strb [8]bool
	Last bool
	User bool
}

//
// Type WriteResp specifies AXI memory write response channel fields.
//
type WriteResp struct {
	Id   bool
	Resp [2]bool
	User bool
}

//
// WriteDisable will disable AXI bus write transactions. Should be run once for each
// unused AXI write interface. This will block the calling goroutine.
//
func WriteDisable(
	clientAddr chan<- Addr,
	clientData chan<- WriteData,
	clientResp <-chan WriteResp) {

	clientAddr <- Addr{}
	clientData <- WriteData{Last: true}
	for {
		<-clientResp
	}
}

//
// ReadDisable will disable AXI bus read transactions. Should be run once for
// each unused AXI read interface. This will block the calling goroutine.
//
func ReadDisable(
	clientAddr chan<- Addr,
	clientData <-chan ReadData) {

	clientAddr <- Addr{}
	for {
		<-clientData
	}
}
