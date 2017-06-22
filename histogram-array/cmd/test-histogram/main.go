package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"xcl"
)

const (
	// The maximum bit width we allow values to have
	MAX_BIT_WIDTH = 16
	// The bit width we will compress to
	HISTOGRAM_BIT_WIDTH = 9
	// The resulting number of elements of the histogram
	HISTOGRAM_WIDTH = 1 << 9
)

func main() {
	// Setup a new world for accessing our kernel
	world := xcl.NewWorld()
	defer world.Release()

	// Import the "kernel_test" file for our FPGA, generated as part
	// of the build process, and from that get the kernel named
	// "reconfigure_io_sdaccel_builder_stub_0_1", which is currently hardcoded
	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	// The data we'll send to the kernel for processing
	input := make([]uint32, 20)

	// seed it with 20 random values, bound to 0 - 2**16
	for i, _ := range input {
		input[i] = uint32(uint16(rand.Uint32()))
	}

	// On the FGPA, allocated ReadOnly memory for the input to the kernel.
	buff := world.Malloc(xcl.ReadOnly, uint(binary.Size(input)))
	defer buff.Free()

	// Construct our local output
	var output [HISTOGRAM_WIDTH]uint32

	// On the FGPA, allocated ReadWrite memory for the output from the kernel.
	outputBuff := world.Malloc(xcl.ReadWrite, uint(binary.Size(output)))
	defer outputBuff.Free()

	// write our input to the kernel at the memory we've previously allocated
	binary.Write(buff.Writer(), binary.LittleEndian, &input)

	// zero out output buffer
	binary.Write(outputBuff.Writer(), binary.LittleEndian, &output)

	// Pass the pointer to the input memory on the FPGA as the first argument
	krnl.SetMemoryArg(0, buff)
	// Pass the pointer to the output memory on the FPGA as the second argument
	krnl.SetMemoryArg(1, outputBuff)
	// Pass the total length of the input as the third argument
	krnl.SetArg(2, uint32(len(input)))

	// Run the kernel
	krnl.Run(1, 1, 1)

	// Read the output from the memory on the FPGA
	err := binary.Read(outputBuff.Reader(), binary.LittleEndian, &output)
	if err != nil {
		log.Fatal("binary.Read failed:", err)
	}

	// Calculate the same values as the kernel did for a test
	var expected [HISTOGRAM_WIDTH]uint32
	for _, val := range input {
		expected[val>>(MAX_BIT_WIDTH-HISTOGRAM_BIT_WIDTH)] += 1
	}

	// error if they didn't do the same calculation
	if !reflect.DeepEqual(expected, output) {
		log.Fatalf("%v != %v\n", output, expected)
	}

	// Print out each bucket, and value
	for i, val := range output {
		fmt.Printf("%d: %d\n", i<<(MAX_BIT_WIDTH-HISTOGRAM_BIT_WIDTH), val)
	}

}
