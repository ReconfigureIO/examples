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

    // Create/get data and pass arguments to the kernel as required. These could be small pieces of data,
    // pointers to memory, data lengths so the Kernel knows what to expect. This all depends on your project.
    // We have passed three arguments here, you can pass more as neccessary

    // make an array to send to the kernel for processing
    input := make([]uint32, 10)

	  // seed it with incrementing values
  	for i, _ := range input {
  		input[i] = uint32(i)
  	}

    // Create space in shared memory for our array input
  	buff := world.Malloc(xcl.ReadOnly, uint(binary.Size(input)))
  	defer buff.Free()

    // Create a variable to hold the output from the FPGA
  	var output [10]uint32

  	// Create space in the shared memory for the output from the FPGA
  	outputBuff := world.Malloc(xcl.ReadWrite, uint(binary.Size(output)))
  	defer outputBuff.Free()

  	// write our input to the shared memory at the location we specified previously
  	binary.Write(buff.Writer(), binary.LittleEndian, &input)

  	// zero out output space
  	binary.Write(outputBuff.Writer(), binary.LittleEndian, &output)

    // Send the location of the input array as the first argument
    krnl.SetMemoryArg(0, buff)
    // Send the location the FPGA should put the result as the second argument
    krnl.SetMemoryArg(1, outputBuff)
    // Send the length of the input array, so the kernel knows what to expect, as the third argument
    krnl.SetArg(2, uint32(len(input)))

    // Run the kernel with the supplied arguments. This is the same for all projects.
    // The arguments ``(1, 1, 1)`` relate to x, y, z co-ordinates and correspond to our current
    // underlying technology.
    krnl.Run(1, 1, 1)

    // Display/use the results returned from the FPGA as required!

    binary.Read(outputBuff.Reader(), binary.LittleEndian, &output);

    for _, val := range output {
      print(val)
    }

}
