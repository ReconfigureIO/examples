package main

import (
	"encoding/binary"
	"fmt"
	"github.com/ReconfigureIO/sdaccel/xcl"
	"os"

	"github.com/ReconfigureIO/fixed"
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

	a := fixed.I26F(2, 1<<5)
	b := fixed.I26(4)

	// Set the first operand
	krnl.SetArg(0, uint32(a))
	// Set the second operand
	krnl.SetArg(1, uint32(b))
	// Set the pointer to the output buffer
	krnl.SetMemoryArg(2, buff)

	// Run the kernel with the supplied arguments
	krnl.Run(1, 1, 1)

	// Decode that byte slice into the uint32 we're expecting
	var ret fixed.Int26_6
	err := binary.Read(buff.Reader(), binary.LittleEndian, &ret)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	expected := a.Mul(b)

	// Exit with an error if the value is not correct
	if expected != ret {
		// Print the value we got from the FPGA
		fmt.Printf("Expected %d, got %d\n", expected, ret)
		os.Exit(1)
	}
}
