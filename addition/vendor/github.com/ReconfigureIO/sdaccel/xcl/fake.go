// +build !opencl

package xcl

import (
	"io"
)

// World is an opaque structure that allows communication with FPGAs.
type World struct {
}

// Program ways to lookup kernels
type Program struct {
	world *World
}

// Kernel is a a function that runs on an FGPA.
type Kernel struct {
	program *Program
}

// Memory represents a segment of RAM on the FGPA
type Memory struct {
	world *World
	size  uint
}

// MemoryWriter is an io.Writer to RAM on the FPGA
type MemoryWriter struct {
	left   uint
	offset uint
	memory *Memory
}

// MemoryReader is an io.Reader to RAM on the FPGA
type MemoryReader struct {
	left   uint
	offset uint
	memory *Memory
}

// Constants for opening RAM on the FGPA
const (
	ReadOnly = iota
	WriteOnly
	ReadWrite
)

/*

NewWorld creates a new World. This needs to be released when done. This can be done using `defer`

    world := xcl.NewWorld()
    defer world.Release()

*/
func NewWorld() World {
	return World{}
}

/*

Release cleans up a previously created World.

*/
func (world *World) Release() {
}

/*

Import will search for an appropriate xclbin and load their contents,
either in a simulator for hardware simulation, or onto an FPGA for
actual hardware. The input argument is the name of the program from
the build procedure (typically "kernel_test"). The returned value is
the program used for interacting with the loaded xclbin.

This needs to be released when done. This can be done using defer.

    program := world.Import("kernel_test")
    defer program.Release()

*/
func (world World) Import(program string) *Program {
	return &Program{&world}
}

/*

GetKernel will return the specific Kernel from the Program. The input
argument is the name of the Kernel in the Program (typically
"reconfigure_io_sdaccel_builder_stub_0_1").

This needs to be released when done.


    kernel := program.GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
    defer kernel.Release()

*/
func (program *Program) GetKernel(kernelName string) *Kernel {
	return &Kernel{program}
}

/*

Release a previously acquired Program.

*/
func (program *Program) Release() {
}

/*

Release a previously acquired Kernel

*/
func (kernel *Kernel) Release() {
}

/*

Malloc allocates a number of bytes on the FPGA. The resulting
structure represents a pointer to Memory on the FGPA.

This needs to be freed when done.

	buff := world.Malloc(xcl.WriteOnly, 512)
	defer buff.Free()

*/
func (world *World) Malloc(flags uint, size uint) *Memory {
	return &Memory{world, size}
}

/*

Free a previously allocated Memory.

*/
func (mem *Memory) Free() {
}

/*

Writer constructs a one-time use writer for a Memory. This has the standard io.Writer interface. For example, to copy data to the FPGA with the binary package:

    var input [256]uint32
	err := binary.Write(buff.Writer(), binary.LittleEndian, &input)

*/
func (mem *Memory) Writer() *MemoryWriter {
	return &MemoryWriter{mem.size, 0, mem}
}

func (writer *MemoryWriter) Write(bytes []byte) (n int, err error) {
	if writer.left == 0 {
		return 0, io.ErrShortWrite
	}
	toWrite := uint(len(bytes))
	if toWrite > writer.left {
		toWrite = writer.left
	}
	writer.left -= toWrite
	writer.offset += toWrite
	return int(toWrite), nil
}

/*

Reader constructs a one-time use reader for a Memory. This has the standard io.Reader interface. For example, to copy from the FPGA with the binary package:

    var input [256]uint32
	err := binary.Read(buff.Reader(), binary.LittleEndian, &input)

*/
func (mem *Memory) Reader() *MemoryReader {
	return &MemoryReader{mem.size, 0, mem}
}

func (reader *MemoryReader) Read(bytes []byte) (n int, err error) {
	if reader.left == 0 {
		return 0, io.EOF
	}
	toRead := uint(len(bytes))
	if toRead > reader.left {
		toRead = reader.left
	}

	reader.left -= toRead
	reader.offset += toRead
	return int(toRead), nil
}

/*

SetMemoryArg passes the pointer to Memory as an argument to the
Kernel. The resulting type on the kernel will be a uintptr.

*/
func (kernel *Kernel) SetMemoryArg(index uint, mem *Memory) {
}

/*

SetArg passes the uint32 as an argument to the Kernel. The resulting
type on the kernel will be a uint32.

*/
func (kernel *Kernel) SetArg(index uint, val uint32) {
}

/*
Run will start execution of the Kernel with the number of dimensions. Most uses of this should be called as

    kernel.Run()

*/
func (kernel *Kernel) Run(_ ...uint) {
}
