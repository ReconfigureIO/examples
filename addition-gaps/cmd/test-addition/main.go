package main

import (
	"encoding/binary"
	"fmt"
	"github.com/ReconfigureIO/sdaccel/xcl"
	"os"
)

func main() {
	// Allocate a world for interacting with kernels
	world := xcl.NewWorld()
	defer world.Release()

	// Import the kernel.
	// Right now these two idenitifers are hard coded as an output from the build process
	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	// Allocate a buffer on the FPGA to store the return value of our computation
	// The output is a uint32, so we need 4 bytes to store it
	buff := world.Malloc(xcl.WriteOnly, 4)
	defer buff.Free()

	// Pass the arguments to the kernel

	// Set the first operand to 1
	// YOUR CODE HERE
	// Set the second operand to 2
	// YOUR CODE HERE
	// Set the pointer to the output buffer
	// YOUR CODE HERE

	// Run the kernel with the supplied arguments
	krnl.Run(1, 1, 1)

	// Decode that byte slice into the uint32 we're expecting
	var ret uint32
	err := binary.Read(buff.Reader(), binary.LittleEndian, &ret)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	// Print the value we got from the FPGA
	// YOUR CODE HERE

	// Exit with an error if the value is not correct
	// YOUR CODE HERE: create an if statment to exit if the value retuned does not equal 3
	//
	//
}
