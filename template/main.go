package main

import (
    // Import the entire framework (including bundled verilog)
    _ "github.com/ReconfigureIO/sdaccel"

    // Use the new AXI protocol package
    aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
    axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
)

func Top(
    // Specify inputs and outputs to the kernel. Tell the kernel where to find data in shared memory, what data type
    // to expect or pass single integers directly to the kernel by sending them to the FPGA's control register

    ...

    // Set up channels for interacting with the shared memory
    memReadAddr chan<- axiprotocol.Addr,
    memReadData <-chan axiprotocol.ReadData,

    memWriteAddr chan<- axiprotocol.Addr,
    memWriteData chan<- axiprotocol.WriteData,
    memWriteResp <-chan axiprotocol.WriteResp) {

    // Do whatever needs doing with the data from the host

    ...

    // Write the result to the location in shared memory as requested by the host
    aximemory.WriteUInt32(
        memWriteAddr, memWriteData, memWriteResp, true, <results_pointer>, <results_data>)
}
