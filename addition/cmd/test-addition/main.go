package main

import (
	"encoding/binary"
	"fmt"
	"github.com/ReconfigureIO/sdaccel/xcl"
	"os"
)

func main() {
	// Allocate a world for interacting with the FPGA
	world := xcl.NewWorld()
	defer world.Release()

	// Import the compiled code that will be loaded onto the FPGA (referred to here as a kernel)
	// Right now these two idenitifers are hard coded as an output from the build process
	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	// Allocate space in shared memory for the FPGA to store the result of the computation
	// The output is a uint32, so we need 4 bytes to store it
	buff := world.Malloc(xcl.WriteOnly, 4)
	defer buff.Free()

	// Pass the arguments to the kernel

	// Set the first operand to 1
	krnl.SetArg(0, 1)
	// Set the second operand to 2
	krnl.SetArg(1, 2)
	// Set the pointer to the result address in shared memory
	krnl.SetMemoryArg(2, buff)

	// Run the FPGA with the supplied arguments. This is the same for all projects.
	// The arguments ``(1, 1, 1)`` relate to x, y, z co-ordinates and correspond to our current
	// underlying technology.
	krnl.Run(1, 1, 1)

	// Create a variable for the result from the FPGA and read the result into it.
	// We have also set an error condition to tell us if the read fails.
	var ret uint32
	err := binary.Read(buff.Reader(), binary.LittleEndian, &ret)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	// Print the value we got from the FPGA
	fmt.Printf("%d\n", ret)

	// Check the result is correct and if not, return an error
	if ret != 3 {
		os.Exit(1)
	}
}
