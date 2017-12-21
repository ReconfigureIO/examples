package main

import (
	"encoding/binary"
	"fmt"
	"github.com/ReconfigureIO/sdaccel/xcl"
)

func main() {
	// Allocate a 'world' for interacting with and FPGA
	world := xcl.NewWorld()
	defer world.Release()

	// Import the compiled code that will be loaded onto the FPGA (referred to here as a kernel)
	// Right now these two identifiers are hard coded as an output from the build process
	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	// Allocate a space in the shared memory to store the results from the FPGA
	// The output is a uint32, so we need 4 bytes to store it
	buff := world.Malloc(xcl.WriteOnly, 4)
	defer buff.Free()

	// Pass the arguments to the kernel.

	// First argument is the integer to be multiplied
	krnl.SetArg(0, 1)
	// Second argument is the pointer to shared memory for storing the result
	krnl.SetMemoryArg(1, buff)

	// Run the FPGA with the supplied arguments. This is the same for all projects.
	// The arguments ``(1, 1, 1)`` relate to x, y, z co-ordinates and correspond to our current
	// underlying technology.
	krnl.Run(1, 1, 1)

	// Create a variable for the result from the FPGA and read the result into it
	var output uint32
	binary.Read(buff.Reader(), binary.LittleEndian, &output)

	// Print the result
	fmt.Printf("%d\n", output)

}
