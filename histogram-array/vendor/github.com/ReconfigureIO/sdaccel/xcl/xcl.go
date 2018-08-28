// +build opencl

// Package xcl provides primitives for working with kernels from the host
package xcl

// #cgo CFLAGS: -std=gnu99
// #cgo LDFLAGS: -lxilinxopencl -llmx6.0
// #include "xcl.h"
// #include <stdlib.h>
//
// cl_int setMemArg(cl_kernel kernel, cl_uint arg_index, cl_mem m) {
//    return clSetKernelArg(kernel, arg_index, sizeof(cl_mem), &m);
// }
//
import "C"

import (
	"errors"
	"fmt"
	"io"
	"unsafe"
)

// World is an opaque structure that allows communication with FPGAs.
type World struct {
	cw C.xcl_world
}

// Program ways to lookup kernels
type Program struct {
	world   *World
	program C.cl_program
}

// Kernel is a a function that runs on an FGPA.
type Kernel struct {
	program *Program
	kernel  C.cl_kernel
}

// Memory represents a segment of RAM on the FGPA
type Memory struct {
	world *World
	size  uint
	mem   C.cl_mem
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
	return World{C.xcl_world_single()}
}

/*

Release cleans up a previously created World.

*/
func (world *World) Release() {
	C.xcl_release_world(world.cw)
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
	s := C.CString(program)
	p := C.xcl_import_binary(world.cw, s)
	C.free(unsafe.Pointer(s))
	return &Program{&world, p}
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
	s := C.CString(kernelName)
	k := C.xcl_get_kernel(C.cl_program(program.program), s)
	C.free(unsafe.Pointer(s))
	return &Kernel{program, k}
}

/*

Release a previously acquired Program.

*/
func (program *Program) Release() {
	C.clReleaseProgram(program.program)
}

/*

Release a previously acquired Kernel

*/
func (kernel *Kernel) Release() {
	C.clReleaseKernel(kernel.kernel)
}

/*

Malloc allocates a number of bytes on the FPGA. The resulting
structure represents a pointer to Memory on the FGPA.

This needs to be freed when done.

	buff := world.Malloc(xcl.WriteOnly, 512)
	defer buff.Free()

*/
func (world *World) Malloc(flags uint, size uint) *Memory {
	var f C.cl_mem_flags
	switch flags {
	case ReadOnly:
		f = C.CL_MEM_READ_ONLY
	case WriteOnly:
		f = C.CL_MEM_WRITE_ONLY
	case ReadWrite:
		f = C.CL_MEM_READ_WRITE
	}
	m := C.xcl_malloc(world.cw, f, C.size_t(size))
	return &Memory{world, size, m}
}

/*

Free a previously allocated Memory.

*/
func (mem *Memory) Free() {
	C.clReleaseMemObject(mem.mem)
}

/*

Writer constructs a one-time use writer for a Memory. This has the standard io.Writer interface. For example, to copy data to the FPGA with the binary package:

    var input [256]uint32
	err := binary.Write(buff.Writer(), binary.LittleEndian, &input)

*/
func (mem *Memory) Writer() *MemoryWriter {
	return &MemoryWriter{mem.size, 0, mem}
}

func errorCode(code C.cl_int) error {
	switch code {
	case C.CL_SUCCESS:
		return nil
	case C.CL_INVALID_COMMAND_QUEUE:
		return errors.New("CL_INVALID_COMMAND_QUEUE")
	case C.CL_INVALID_CONTEXT:
		return errors.New("CL_INVALID_CONTEXT")
	case C.CL_INVALID_MEM_OBJECT:
		return errors.New("CL_INVALID_MEM_OBJECT")
	case C.CL_INVALID_VALUE:
		return errors.New("CL_INVALID_VALUE")
	case C.CL_INVALID_EVENT_WAIT_LIST:
		return errors.New("CL_INVALID_EVENT_WAIT_LIST")
	case C.CL_MISALIGNED_SUB_BUFFER_OFFSET:
		return errors.New("CL_MISALIGNED_SUB_BUFFER_OFFSET")
	case C.CL_EXEC_STATUS_ERROR_FOR_EVENTS_IN_WAIT_LIST:
		return errors.New("CL_EXEC_STATUS_ERROR_FOR_EVENTS_IN_WAIT_LIST")
	case C.CL_MEM_OBJECT_ALLOCATION_FAILURE:
		return errors.New("CL_MEM_OBJECT_ALLOCATION_FAILURE")
	case C.CL_INVALID_OPERATION:
		return errors.New("CL_INVALID_OPERATION")
	case C.CL_OUT_OF_RESOURCES:
		return errors.New("CL_OUT_OF_RESOURCES")
	case C.CL_OUT_OF_HOST_MEMORY:
		return errors.New("CL_OUT_OF_HOST_MEMORY")
	default:
		return fmt.Errorf("Unknown error code %d", code)
	}
}

func (writer *MemoryWriter) Write(bytes []byte) (n int, err error) {
	if writer.left == 0 {
		return 0, io.ErrShortWrite
	}
	toWrite := uint(len(bytes))
	if toWrite > writer.left {
		toWrite = writer.left
	}
	// I think we can make this zero copy like in Read
	p := C.CBytes(bytes[0:toWrite])

	ret := C.clEnqueueWriteBuffer(
		writer.memory.world.cw.command_queue,
		writer.memory.mem,
		C.CL_TRUE,
		C.size_t(writer.offset), C.size_t(toWrite), p, C.cl_uint(0), nil, nil)

	err = errorCode(ret)
	C.free(p)
	writer.left -= toWrite
	writer.offset += toWrite
	return int(toWrite), err
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

	p := unsafe.Pointer(&bytes[0])

	ret := C.clEnqueueReadBuffer(
		reader.memory.world.cw.command_queue,
		reader.memory.mem,
		C.CL_TRUE,
		C.size_t(reader.offset), C.size_t(toRead), p, C.cl_uint(0), nil, nil)

	err = errorCode(ret)
	reader.left -= toRead
	reader.offset += toRead
	return int(toRead), err
}

/*

SetMemoryArg passes the pointer to Memory as an argument to the
Kernel. The resulting type on the kernel will be a uintptr.

*/
func (kernel *Kernel) SetMemoryArg(index uint, mem *Memory) {
	C.setMemArg(kernel.kernel, C.cl_uint(index), mem.mem)
}

/*

SetArg passes the uint32 as an argument to the Kernel. The resulting
type on the kernel will be a uint32.

*/
func (kernel *Kernel) SetArg(index uint, val uint32) {
	C.clSetKernelArg(kernel.kernel, C.cl_uint(index), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val))
}

/*
Run will start execution of the Kernel. Most uses of this should be called as

    kernel.Run()

*/
func (kernel *Kernel) Run(_ ...uint) error {
	size := C.size_t(1)
	event := new(C.cl_event)

	errCode := C.clEnqueueNDRangeKernel(kernel.program.world.cw.command_queue, kernel.kernel, 1,
		nil, &size, &size, 0, nil, event)
	err := errorCode(errCode)
	if err != nil {
		return err
	}

	errCode = C.clFlush(kernel.program.world.cw.command_queue)
	err = errorCode(errCode)
	if err != nil {
		return err
	}

	errCode = C.clWaitForEvents(1, event)
	err = errorCode(errCode)
	return err
}
