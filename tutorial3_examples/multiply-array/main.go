package main

import (
    // Import the entire framework (including bundled verilog)
    _ "sdaccel"

    // Use the new AXI protocol package
    aximemory "axi/memory"
    axiprotocol "axi/protocol"
)

func Top(
    // Specify inputs and outputs to the kernel. Tell the kernel where to find data in shared memory, what data type
    // to expect or pass single integers directly to the kernel by sending them to the FPGA's control register

    inputData uintptr,
    outputData uintptr,
    length uint32,

    // Set up channels for interacting with the shared memory
    memReadAddr chan<- axiprotocol.Addr,
    memReadData <-chan axiprotocol.ReadData,

    memWriteAddr chan<- axiprotocol.Addr,
    memWriteData chan<- axiprotocol.WriteData,
    memWriteResp <-chan axiprotocol.WriteResp) {

    // Do whatever needs doing with the data from the host

    // Read all the input data into a channel
    inputChan := make(chan uint32)
    go aximemory.ReadBurstUInt32(
        memReadAddr, memReadData, true, inputData, length, inputChan)

    // Create a channel for the result of the calculation
    transformedChan := make(chan uint32)
    // multiply each element of the input channel by 2 and send to the channel we just made to hold the result
    go func(){
        // no need to stop here, which will save us some clocks checking
        for {
            transformedChan <- (<-inputChan) * 2
        }
    }()
    
    // Write transformed results back to memory
    aximemory.WriteBurstUInt32(
        memWriteAddr, memWriteData, memWriteResp, true, outputData, length, transformedChan)
}
