package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"reflect"

	"github.com/ReconfigureIO/sdaccel/xcl"
)

// Define constants to be used in func main
const (
	// The maximum bit width we allow values to have
	MAX_BIT_WIDTH = 16
	// The bit width we will compress to
	HISTOGRAM_BIT_WIDTH = 9
	// The resulting number of elements of the histogram
	HISTOGRAM_WIDTH = 1 << 9
)

func main() {
	// Allocate a 'world' for interacting with the FPGA
	world := xcl.NewWorld()
	defer world.Release()

	// Import the compiled code that will be loaded onto the FPGA (referred to here as a kernel)
	// Right now these two identifiers are hard coded as an output from the build process
	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	// Define a new array for the data we'll send to the FPGA for processing
	input := make([]uint32, 20)

	// Seed it with 20 random values, bound to 0 - 2**16
	for i, _ := range input {
		input[i] = uint32(uint16(rand.Uint32()))
	}

	// Allocate a space in the shared memory to store the data you're sending to the FPGA
	buff := world.Malloc(xcl.ReadOnly, uint(binary.Size(input)))
	defer buff.Free()

	// Construct an array to hold the output data from the FPGA
	var output [HISTOGRAM_WIDTH]uint32

	// Allocate a space in the shared memory to store the output data from the FPGA
	outputBuff := world.Malloc(xcl.ReadWrite, uint(binary.Size(output)))
	defer outputBuff.Free()

	// Write our input data to shared memory at the address we previously allocated
	binary.Write(buff.Writer(), binary.LittleEndian, &input)

	// Zero out the space in shared memory for the result from the FPGA
	binary.Write(outputBuff.Writer(), binary.LittleEndian, &output)

	// Pass the pointer to the input data in shared memory as the first argument
	krnl.SetMemoryArg(0, buff)
	// Pass the pointer to the memory location reserved for the result as the second argument
	krnl.SetMemoryArg(1, outputBuff)
	// Pass the total length of the input as the third argument
	krnl.SetArg(2, uint32(len(input)))

	// Run the FPGA with the supplied arguments. This is the same for all projects.
	// The arguments ``(1, 1, 1)`` relate to x, y, z co-ordinates and correspond to our current
	// underlying technology.
	krnl.Run(1, 1, 1)

	// Read the result from shared memory. If it is zero return an error
	err := binary.Read(outputBuff.Reader(), binary.LittleEndian, &output)
	if err != nil {
		log.Fatal("binary.Read failed:", err)
	}

	// Calculate the same values locally to check the FPGA got it right
	var expected [HISTOGRAM_WIDTH]uint32
	for _, val := range input {
		expected[val>>(MAX_BIT_WIDTH-HISTOGRAM_BIT_WIDTH)] += 1
	}

	// Return an error if the local and FPGA calculations do not give the same result
	if !reflect.DeepEqual(expected, output) {
		log.Fatalf("%v != %v\n", output, expected)
	}

	log.Println()
	log.Printf("We sent an array of 20 integers to the FPGA for processing: \n")
	log.Printf("Input: %v \n", input)
	log.Printf("We programmed the FPGA to sort the data into bins, and these are the results we got: \n")

	// Print out each bin and coresponding value
	for i, val := range output {
		fmt.Printf("%d: %d\n", i<<(MAX_BIT_WIDTH-HISTOGRAM_BIT_WIDTH), val)
	}

}
