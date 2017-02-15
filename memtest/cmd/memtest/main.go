package main

import (
	"log"
	"xcl"
)

func main() {
	world := xcl.NewWorld()
	defer world.Release()

	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	byteLength := uint(64)

	outputBuff := world.Malloc(xcl.WriteOnly, byteLength)
	defer outputBuff.Free()

	krnl.SetMemoryArg(0, outputBuff)

	krnl.Run(1, 1, 1)

	resp := make([]byte, byteLength)
	outputBuff.Read(resp)

	expected := make([]byte, byteLength)

	for i := 0; i < int(byteLength); i++ {
		expected[i] = 1
	}

	log.Printf("%v == %v", resp, expected)
}
