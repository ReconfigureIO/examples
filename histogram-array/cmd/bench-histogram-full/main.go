package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/ReconfigureIO/sdaccel/xcl"
)

func main() {
	// take the first command line argument and use as the data size for the benchmark
	input := os.Args[1]

	// convert the string argument to an int
	nInputs, err := strconv.Atoi(input)
	if err != nil {
		// handle error
		fmt.Println(err)
		os.Exit(2)
	}

	// initialise a new state using our specified input size and warm up
	state := NewState(nInputs)
	defer state.Release()

	// run the benchmark
	log.Println()
	log.Println()
	log.Printf("Time taken to pass, process and collect an array of %v integers: \n", nInputs)
	log.Println()

	result := testing.Benchmark(state.Run)
	fmt.Println(result)
}

type State struct {
	// Everything that needs setting up - kernel, input buffer, output buffer, input var, result var.
	world      xcl.World
	program    *xcl.Program
	krnl       *xcl.Kernel
	inputBuff  *xcl.Memory
	outputBuff *xcl.Memory
	input      []uint32
	output     []uint32
}

func NewState(nInputs int) *State {
	w := xcl.NewWorld()          // variable for new World
	p := w.Import("kernel_test") // variable to import our kernel
	size := uint(nInputs) * 4    // number of bytes needed to hold the input and output data

	s := &State{
		world:      w,                                                      // allocate a new world for interacting with the FPGA
		program:    p,                                                      // Import the compiled code that will be loaded onto the FPGA (referred to here as a kernel)
		krnl:       p.GetKernel("reconfigure_io_sdaccel_builder_stub_0_1"), // Right now these two identifiers are hard coded as an output from the build process
		inputBuff:  w.Malloc(xcl.ReadOnly, size),                           // constructed an input buffer as a function of nInputs
		outputBuff: w.Malloc(xcl.ReadWrite, size),                          // In this example our output will be the same size as our input
		input:      make([]uint32, nInputs),                                // make a variable to store our input data
		output:     make([]uint32, nInputs),                                // make a variable to store our results data
	}

	// Seed the input array with random values
	for i, _ := range s.input {
		s.input[i] = uint32(uint16(rand.Uint32()))
	}

	//To avoid measuring warmup cost of the first few calls (especially in sim)
	const warmup = 2
	for i := 0; i < warmup; i++ {
		s.feedFPGA()
	}

	return s
}

// This function will calculate the benchmark, it will run repeatedly until it achieves a reliable result
func (s *State) Run(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s.feedFPGA()
	}
}

// This function frees up buffers and released the World an program used to interact with the FPGA
func (s *State) Release() {
	s.inputBuff.Free()
	s.outputBuff.Free()
	s.program.Release()
	s.world.Release()
}

// This function writes our sample data to memory, tells the FPGA where it is, and where to put the result and starts the FPGA runnings
func (s *State) feedFPGA() {
	// write input to memory
	binary.Write(s.inputBuff.Writer(), binary.LittleEndian, &s.input)

	s.krnl.SetMemoryArg(0, s.inputBuff)    // Send the location of the input data as the first argument
	s.krnl.SetMemoryArg(1, s.outputBuff)   // Send the location the FPGA should put the result as the second argument
	s.krnl.SetArg(2, uint32(len(s.input))) // Send the length of the input array as the third argument, so the FPGA knows what to expect

	// start the FPGA running
	s.krnl.Run(1, 1, 1)

	// Read the results into our output variable
	binary.Read(s.outputBuff.Reader(), binary.LittleEndian, &s.output)

	// log.Printf("Input: %v ", s.input)
	// log.Printf("Output: %v ", s.output)
}
