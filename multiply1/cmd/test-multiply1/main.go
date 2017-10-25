package main

import (
    "encoding/binary"
    "xcl"
    "fmt"
)

func main() {
    // Allocate a 'world' for interacting with kernels
    world := xcl.NewWorld()
    defer world.Release()

    // Import the kernel.
    // Right now these two identifiers are hard coded as an output from the build process
    krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
    defer krnl.Release()

    // Allocate a space in the shared memory to store the results from the FPGA
    // The output is a uint32, so we need 4 bytes to store it
    buff := world.Malloc(xcl.WriteOnly, 4)
    defer buff.Free()

    // Pass the arguments to the kernel. These could be small pieces of data, pointers to
    // memory, data lengths so the Kernel knows what to expect. This all depends on your project.

    // First argument
    krnl.SetArg(0, 1)
    // Second argument
    krnl.SetMemoryArg(1, buff)

    // Run the kernel with the supplied arguments. This is the same for all projects.
    // The arguments ``(1, 1, 1)`` relate to x, y, z co-ordinates and correspond to our current
    // underlying technology.
    krnl.Run(1, 1, 1)

    // Display/use the results returned from the FPGA as required!
  	var output uint32
  	binary.Read(outputBuff.Reader(), binary.LittleEndian, &output);

  	// Print the value we got from the FPGA
  	fmt.Printf("%d\n", output)

}
