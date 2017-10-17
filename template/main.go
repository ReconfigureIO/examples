package main

import (
)

func Top(
    // Specify inputs and outputs to the kernel. Tell the kernel where to find data in shared memory, what data type
    // to expect or pass single integers directly to the kernel by sending them to the FPGA's control register

    ...

    // Next set up channels for interacting with the shared memory
    memReadAddr chan<- axiprotocol.Addr,
    memReadData <-chan axiprotocol.ReadData,

    memWriteAddr chan<- axiprotocol.Addr,
    memWriteData chan<- axiprotocol.WriteData,
    memWriteResp <-chan axiprotocol.WriteResp) {

    // Do whatever needs doing with the data from the host

    ...

    // Write the result to the location in shared memory dictated by the host
    aximemory.WriteUInt32(
        memWriteAddr, memWriteData, memWriteResp, false, addr, uint32(val))
}
