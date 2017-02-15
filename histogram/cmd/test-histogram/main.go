package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"xcl"
)

const (
	MAX_BIT_WIDTH       = 16
	HISTOGRAM_BIT_WIDTH = 9
	HISTOGRAM_WIDTH     = 1 << 9
)

func main() {
	world := xcl.NewWorld()
	defer world.Release()

	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	input := make([]uint32, 20)

	// seed it with 20 random values, bound to 0 - 2**16
	for i, _ := range input {
		input[i] = uint32(uint16(rand.Uint32()))
	}

	inputByteLength := uint(4 * len(input))

	buff := world.Malloc(xcl.ReadOnly, inputByteLength)
	defer buff.Free()

	resp := make([]byte, 4*HISTOGRAM_WIDTH)
	outputBuff := world.Malloc(xcl.ReadWrite, uint(len(resp)))
	defer outputBuff.Free()

	inputBuff := new(bytes.Buffer)
	binary.Write(inputBuff, binary.LittleEndian, &input)
	buff.Write(inputBuff.Bytes())

	outputBuff.Write(resp)

	krnl.SetMemoryArg(0, buff)
	krnl.SetMemoryArg(1, outputBuff)
	krnl.SetArg(2, uint32(len(input)))

	krnl.Run(1, 1, 1)

	outputBuff.Read(resp)

	var ret [512]uint32
	err := binary.Read(bytes.NewReader(resp), binary.LittleEndian, &ret)
	if err != nil {
		log.Fatal("binary.Read failed:", err)
	}

	var expected [HISTOGRAM_WIDTH]uint32

	for _, val := range input {
		expected[val>>(MAX_BIT_WIDTH-HISTOGRAM_BIT_WIDTH)] += 1
	}

	if !reflect.DeepEqual(expected, ret) {
		log.Fatalf("%v != %v\n", ret, expected)
	}

	for i, val := range ret {
		fmt.Printf("%d: %d\n", i<<(MAX_BIT_WIDTH-HISTOGRAM_BIT_WIDTH), val)
	}

}
