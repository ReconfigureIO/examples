package main

import (
	"encoding/binary"
	"log"

	"github.com/ReconfigureIO/sdaccel/xcl"
)

const DATA_WIDTH = 1536
const ITERATIONS = 2

func main() {
	world := xcl.NewWorld()
	defer world.Release()

	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	inputBuff := world.Malloc(xcl.WriteOnly, DATA_WIDTH)
	defer inputBuff.Free()

	var errResult uint64
	var dcountResult uint64

	errOutBuff := world.Malloc(xcl.WriteOnly, uint(binary.Size(errResult)))
	defer errOutBuff.Free()

	dcountOutBuff := world.Malloc(xcl.WriteOnly, uint(binary.Size(dcountResult)))
	defer dcountOutBuff.Free()

	burstCount := uint32(ITERATIONS)

	krnl.SetMemoryArg(0, inputBuff)
	krnl.SetArg(1, DATA_WIDTH)
	krnl.SetArg(2, burstCount)
	krnl.SetMemoryArg(3, dcountOutBuff)
	krnl.SetMemoryArg(4, errOutBuff)

	krnl.Run(1, 1, 1)

	err := binary.Read(errOutBuff.Reader(), binary.LittleEndian, &errResult)
	if err != nil {
		log.Fatal("binary.Read failed:", err)
	}

	err = binary.Read(dcountOutBuff.Reader(), binary.LittleEndian, &dcountResult)
	if err != nil {
		log.Fatal("binary.Read failed:", err)
	}

	log.Printf("Read %d bytes with %d errors", dcountResult, errResult)
	if errResult != 0 {
		log.Fatal("Read/write errors detected.")
	}
}
