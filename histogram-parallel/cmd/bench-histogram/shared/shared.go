package shared

import (
	"encoding/binary"
	"log"
	"math/rand"
	"testing"
	"github.com/ReconfigureIO/sdaccel/xcl"

	"ReconfigureIO/reco-sdaccel/benchmarks"
)

const (
	MAX_BIT_WIDTH       = 16
	HISTOGRAM_BIT_WIDTH = 9
	HISTOGRAM_WIDTH     = 1 << 9
)

func Process(name string) {
	world := xcl.NewWorld()
	defer world.Release()

	f := func(B *testing.B) {
		program := world.Import("kernel_test")
		defer program.Release()

		krnl := program.GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
		defer krnl.Release()

		doit(world, krnl, B)
	}

	bm := testing.Benchmark(f)
	benchmarks.GipedaResults(name, bm)
}

func doit(world xcl.World, krnl *xcl.Kernel, B *testing.B) {
	log.Printf("Running with N: %d", B.N)
	B.SetBytes(4)
	B.ReportAllocs()

	input := make([]uint32, B.N)

	// seed it with 20 random values, bound to 0 - 2**16
	for i, _ := range input {
		input[i] = uint32(uint16(rand.Uint32()))
	}

	buff := world.Malloc(xcl.ReadOnly, uint(binary.Size(input)))
	defer buff.Free()

	resp := make([]byte, 4*HISTOGRAM_WIDTH)
	outputBuff := world.Malloc(xcl.ReadWrite, uint(binary.Size(resp)))
	defer outputBuff.Free()

	binary.Write(buff.Writer(), binary.LittleEndian, &input)
	log.Printf("Wrote input of size: %d", uint(binary.Size(input)))
	binary.Write(outputBuff.Writer(), binary.LittleEndian, &resp)

	krnl.SetMemoryArg(0, buff)
	log.Printf("Set arg 1")
	krnl.SetMemoryArg(1, outputBuff)
	log.Printf("Set arg 2")
	krnl.SetArg(2, uint32(len(input)))
	log.Printf("Set arg 3")

	log.Printf("Run")
	B.ResetTimer()
	krnl.Run(1, 1, 1)
	B.StopTimer()
	log.Printf("Done")
}
