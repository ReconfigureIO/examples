package main

import (
    "os"
    "xcl"
)

func main() {
    // Allocate a 'world' for interacting with kernels
    world := xcl.NewWorld()
    defer world.Release()

    // Import the kernel.
    // Right now these two identifiers are hard coded as an output from the build process
    krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
    defer krnl.Release()

    // Allocate a space in the shared memory to store the data you're sending to the FPGA and space
    // for the results from the FPGA
    inputBuff := world.Malloc(xcl.ReadOnly, <size here>)
	  defer inputBuff.Free()

    outputBuff := world.Malloc(xcl.WriteOnly, <size here>)
    defer outputBuff.Free()

    // Create/get data and pass arguments to the kernel as required. These could be small pieces of data,
    // pointers to memory, data lengths so the Kernel knows what to expect. This all depends on your project.
    // Usually, you will send data via shared memory, so you will need to write it to the space you allocated
    // above before passing the pointer to the kernel
    // We have passed three arguments here, you can pass more as neccessary

    // First argument
    krnl.SetArg(0, <first>)
    // Second argument
    krnl.SetArg(1, <second>)
    // Third argument
    krnl.SetMemoryArg(2, <third>)

    // Run the kernel with the supplied arguments. This is the same for all projects.
    // The arguments ``(1, 1, 1)`` relate to x, y, z co-ordinates and correspond to our current
    // underlying technology.
    krnl.Run(1, 1, 1)

    // Display/use the results returned from the FPGA as required!

    ...

}
