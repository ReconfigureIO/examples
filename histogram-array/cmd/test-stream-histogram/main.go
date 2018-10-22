package main

import (
	"encoding/binary"
	"log"
	"math/rand"

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

// An example of streaming data while the FPGA is running
func main() {
	// Allocate a 'world' for interacting with the FPGA
	world := xcl.NewWorld()
	defer world.Release()

	// Import the compiled code that will be loaded onto the FPGA (referred to here as a kernel)
	// Right now these two identifiers are hard coded as an output from the build process
	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	// This is for soft guarantees. We can feasibly have
	// (CONCURRENT_RUNS * 2 + 3) separate runs in flight using this
	// scheme. Ideally, we'd provide better guarantees, but that
	// requires a more complex data structure.

	const (
		RUNS            = 10
		CONCURRENT_RUNS = 4
	)

	// This chan is the enqueued runs
	runs := make(chan Histogram, CONCURRENT_RUNS)

	// This chan is for finished, but unprocessed runs
	finished := make(chan Histogram, CONCURRENT_RUNS)

	// Enqueue all the runs we want
	go func() {
		for i := 0; i < RUNS; i++ {
			// Define a new array for the data we'll send to the FPGA for processing
			input := make([]uint32, 20)

			// Seed it with 20 random values, bound to 0 - 2**16
			for i := range input {
				input[i] = uint32(uint16(rand.Uint32()))
			}

			// Here we transfer data over
			h := New(i, world, input)
			runs <- h
		}
		close(runs)
	}()

	// A separate goroutine to run every Histogram on the FPGA
	go func() {
		for histogram := range runs {
			histogram.Run(krnl)
			finished <- histogram
		}
		close(finished)
	}()

	// A final sink to pull in all the finished results
	for histogram := range finished {
		// For convenience, this frees the memory
		_, err := histogram.Result()
		if err != nil {
			log.Println("Error for histogram: ", err)
		}
		// do something with the output

	}
}

type Histogram struct {
	outputBuff *xcl.Memory
	inputBuff  *xcl.Memory
	length     int
	id         int
}

// Create a new Histogram we can run as a Kernel
func New(id int, world xcl.World, input []uint32) Histogram {
	// Construct an array to hold the output data from the FPGA
	var output [HISTOGRAM_WIDTH]uint32

	// Allocate a space in the shared memory to store the data you're sending to the FPGA
	buff := world.Malloc(xcl.ReadOnly, uint(binary.Size(input)))

	// Allocate a space in the shared memory to store the output data from the FPGA
	outputBuff := world.Malloc(xcl.ReadWrite, uint(binary.Size(output)))

	log.Printf("Copying histogram %d", id)

	// Write our input data to shared memory at the address we previously allocated
	binary.Write(buff.Writer(), binary.LittleEndian, &input)

	// Zero out the space in shared memory for the result from the FPGA
	binary.Write(outputBuff.Writer(), binary.LittleEndian, &output)

	return Histogram{
		outputBuff: outputBuff,
		inputBuff:  buff,
		length:     len(input),
		id:         id,
	}
}

func (h Histogram) Run(krnl *xcl.Kernel) {
	log.Printf("Running histogram %d", h.id)

	// Pass the pointer to the input data in shared memory as the first argument
	krnl.SetMemoryArg(0, h.inputBuff)
	// Pass the pointer to the memory location reserved for the result as the second argument
	krnl.SetMemoryArg(1, h.outputBuff)
	// Pass the total length of the input as the third argument
	krnl.SetArg(2, uint32(h.length))

	// Run the FPGA with the supplied arguments. This is the same for all projects.
	// The arguments ``(1, 1, 1)`` relate to x, y, z co-ordinates and correspond to our current
	// underlying technology.
	krnl.Run(1, 1, 1)
}

func (h Histogram) Result() ([HISTOGRAM_WIDTH]uint32, error) {
	log.Printf("Retrieving histogram %d", h.id)

	var output [HISTOGRAM_WIDTH]uint32

	// Read the result from shared memory. If it is zero return an error
	err := binary.Read(h.outputBuff.Reader(), binary.LittleEndian, &output)

	// Free our allocated memory on the FPGA
	h.inputBuff.Free()
	h.outputBuff.Free()

	return output, err
}
