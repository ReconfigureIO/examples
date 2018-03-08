// +build opencl

/**********
Copyright (c) 2016, Xilinx, Inc.
All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
may be used to endorse or promote products derived from this software
without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO,
THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE,
EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
**********/


#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <limits.h>
#include <sys/stat.h>

#include <string.h>
#include <math.h>

#include "xcl.h"

static void* smalloc(size_t size) {
	void* ptr;

	ptr = malloc(size);

	if (ptr == NULL) {
		printf("Error: Cannot allocate memory\n");
		exit(EXIT_FAILURE);
	}

	return ptr;
}

static int load_file_to_memory(const char *filename, char **result) {
	unsigned int size;

	FILE *f = fopen(filename, "rb");
	if (f == NULL) {
		*result = NULL;
		printf("Error: Could not read file %s\n", filename);
		exit(EXIT_FAILURE);
	}

	fseek(f, 0, SEEK_END);
	size = ftell(f);
	fseek(f, 0, SEEK_SET);

	*result = (char *) smalloc(sizeof(char)*(size+1));

	if (size != fread(*result, sizeof(char), size, f)) {
		free(*result);
		printf("Error: read of kernel failed\n");
		exit(EXIT_FAILURE);
	}

	fclose(f);
	(*result)[size] = 0;

	return size;
}

char* xcl_create_and_set(const char* str) {
	size_t len = strlen(str);
	char *ret = (char*) malloc(sizeof(char)*len);
	if (ret == NULL) {
		printf("ERROR: Out of Memory\n");
		exit(EXIT_FAILURE);
	}
	strcpy(ret, str);

	return ret;
}

xcl_world xcl_world_single() {
	int err;
	xcl_world world;
	cl_uint num_platforms;

	char *xcl_mode = getenv("XCL_EMULATION_MODE");
	char *xcl_target = getenv("XCL_TARGET");

	if(xcl_mode == NULL) {
		world.mode = xcl_create_and_set("hw");
	} else {
		/* if xcl_mode is set then check if it's equal to true*/
		if(strcmp(xcl_mode,"true") == 0) {
			/* if it's true, then check if xcl_target is set */
			if(xcl_target == NULL) {
				/* default if emulation but not specified is software emulation */
				world.mode = xcl_create_and_set("sw_emu");
			} else {
				/* otherwise, it's what ever is specified in XCL_TARGET */
				world.mode = xcl_create_and_set(xcl_target);
			}
		} else {
			/* if it's not equal to true then it should be whatever
			 * XCL_EMULATION_MODE is set to */
			world.mode = xcl_create_and_set(xcl_mode);
		}

		err = setenv("XCL_EMULATION_MODE", "true", 1);
		if(err != 0) {
			printf("Error: cannot set XCL_EMULATION_MODE\n");
			exit(EXIT_FAILURE);
		}
	}

	err = clGetPlatformIDs(0, NULL, &num_platforms);
	if (err != CL_SUCCESS) {
		printf("Error: no platforms available or OpenCL install broken\n");
		exit(EXIT_FAILURE);
	}

	cl_platform_id *platform_ids = (cl_platform_id *) malloc(sizeof(cl_platform_id) * num_platforms);

	if (platform_ids == NULL) {
		printf("Error: Out of Memory\n");
		exit(EXIT_FAILURE);
	}

	err = clGetPlatformIDs(num_platforms, platform_ids, NULL);
	if (err != CL_SUCCESS) {
		printf("Error: Failed to find an OpenCL platform!\n");
		exit(EXIT_FAILURE);
	}

	size_t i;
	for(i = 0; i < num_platforms; i++) {
		size_t platform_name_size;
		err = clGetPlatformInfo(platform_ids[i], CL_PLATFORM_NAME,
		                        0, NULL, &platform_name_size);
		if( err != CL_SUCCESS) {
			printf("Error: Could not determine platform name!\n");
			exit(EXIT_FAILURE);
		}

		char *platform_name = (char*) malloc(sizeof(char)*platform_name_size);
		if(platform_name == NULL) {
			printf("Error: out of memory!\n");
			exit(EXIT_FAILURE);
		}

		err = clGetPlatformInfo(platform_ids[i], CL_PLATFORM_NAME,
		                        platform_name_size, platform_name, NULL);
		if(err != CL_SUCCESS) {
			printf("Error: could not determine platform name!\n");
			exit(EXIT_FAILURE);
		}

		if (!strcmp(platform_name, "Xilinx")) {
			free(platform_name);
			world.platform_id = platform_ids[i];
			break;
		}

		free(platform_name);
	}

	free(platform_ids);

	if (i == num_platforms) {
		printf("Error: Failed to find Xilinx platform\n");
		exit(EXIT_FAILURE);
	}

	err = clGetDeviceIDs(world.platform_id, CL_DEVICE_TYPE_ALL,
	                     1, &world.device_id, NULL);
	if (err != CL_SUCCESS) {
		printf("Error: could not get device ids\n");
		exit(EXIT_FAILURE);
	}

	size_t device_name_size;
	err = clGetDeviceInfo(world.device_id, CL_DEVICE_NAME,
	                      0, NULL, &device_name_size);
	if(err != CL_SUCCESS) {
		printf("Error: could not determine device name\n");
		exit(EXIT_FAILURE);
	}

	world.device_name = (char*) malloc(sizeof(char)*device_name_size);

	if(world.device_name == NULL) {
		printf("Error: Out of Memory!\n");
		exit(EXIT_FAILURE);
	}

	err = clGetDeviceInfo(world.device_id, CL_DEVICE_NAME,
	                      device_name_size, world.device_name, NULL);
	if(err != CL_SUCCESS) {
		printf("Error: could not determine device name\n");
		exit(EXIT_FAILURE);
	}

	world.context = clCreateContext(0, 1, &world.device_id,
	                                NULL, NULL, &err);
	if (err != CL_SUCCESS) {
		printf("Error: Failed to create a compute context!\n");
		exit(EXIT_FAILURE);
	}

	world.command_queue = clCreateCommandQueue(world.context,
	                                           world.device_id,
	                                           CL_QUEUE_PROFILING_ENABLE,
	                                           &err);
	if (err != CL_SUCCESS) {
		printf("Error: Failed to create a command queue!\n");
		exit(EXIT_FAILURE);
	}

	return world;
}

void xcl_release_world(xcl_world world) {
	clReleaseCommandQueue(world.command_queue);
	clReleaseContext(world.context);
	free(world.device_name);
	free(world.mode);
}

cl_program xcl_import_binary_file(xcl_world world,
                            const char *xclbin_file_name
) {
	int err;

	if(access(xclbin_file_name, R_OK) != 0) {
		printf("ERROR: %s xclbin not available please build\n", xclbin_file_name);
		exit(EXIT_FAILURE);
	}

	char *krnl_bin;
	const size_t krnl_size = load_file_to_memory(xclbin_file_name, &krnl_bin);

	cl_program program = clCreateProgramWithBinary(world.context, 1,
	                                    &world.device_id, &krnl_size,
	                                    (const unsigned char**) &krnl_bin,
	                                    NULL, &err);
	if ((!program) || (err!=CL_SUCCESS)) {
		printf("Error: Failed to create compute program from binary %d!\n",
		       err);
		printf("Test failed\n");
		exit(EXIT_FAILURE);
	}

	err = clBuildProgram(program, 0, NULL, NULL, NULL, NULL);
	if (err != CL_SUCCESS) {
		size_t len;
		char buffer[2048];

		clGetProgramBuildInfo(program, world.device_id, CL_PROGRAM_BUILD_LOG,
		                      sizeof(buffer), buffer, &len);
		printf("%s\n", buffer);
		printf("Error: Failed to build program executable!\n");
		exit(EXIT_FAILURE);
	}

	free(krnl_bin);

	return program;
}

cl_program xcl_import_binary(xcl_world world,
                            const char *xclbin_name
) {
    char *xcl_bindir = getenv("XCL_BINDIR");

    // typical locations of directory containing xclbin files
    const char *dirs[] = {
        xcl_bindir, // $XCL_BINDIR-specified
        "xclbin",   // command line build
        "..",       // gui build + run
        ".",        // gui build, run in build directory
        NULL
    };
    const char **search_dirs = dirs;
    if (xcl_bindir == NULL) {
    	search_dirs++;
    }

    char *device_name = strdup(world.device_name);
    if (device_name == NULL) {
        printf("Error: Out of Memory\n");
        exit(EXIT_FAILURE);
    }

    // fix up device name to avoid colons and dots.
    // xilinx:xil-accel-rd-ku115:4ddr-xpr:3.2 -> xilinx_xil-accel-rd-ku115_4ddr-xpr_3_2
    for (char *c = device_name; *c != 0; c++) {
        if (*c == ':' || *c == '.') {
            *c = '_';
        }
    }

    const char *file_patterns[] = {
        "%1$s/%2$s.%3$s.%4$s.xclbin",     // <kernel>.<target>.<device>.xclbin
        "%1$s/binary_container_1.xclbin", // default for gui projects
        "%1$s/%2$s.xclbin",               // <kernel>.xclbin
        NULL
    };
    char xclbin_file_name[PATH_MAX];
    memset(xclbin_file_name, 0, PATH_MAX);
    ino_t ino = 0; // used to avoid errors if an xclbin found via multiple/repeated paths
    for (const char **dir = search_dirs; *dir != NULL; dir++) {
        struct stat sb;
        if (stat(*dir, &sb) == 0 && S_ISDIR(sb.st_mode)) {
            for (const char **pattern = file_patterns; *pattern != NULL; pattern++) {
                char file_name[PATH_MAX];
                memset(file_name, 0, PATH_MAX);
                snprintf(file_name, PATH_MAX, *pattern, *dir, xclbin_name, world.mode, device_name);
                if (stat(file_name, &sb) == 0 && S_ISREG(sb.st_mode)) {
                    world.bindir = strdup(*dir);
                    if (world.bindir == NULL) {
                        printf("Error: Out of Memory\n");
                        exit(EXIT_FAILURE);
                    }
                    if (*xclbin_file_name && sb.st_ino != ino) {
                    	printf("Error: multiple xclbin files discovered:\n %s\n %s\n", file_name, xclbin_file_name);
                    	exit(EXIT_FAILURE);
                    }
                    ino = sb.st_ino;
                    strncpy(xclbin_file_name, file_name, PATH_MAX);
                }
            }
        }
    }
	// if no xclbin found, preferred path for error message from xcl_import_binary_file()
    if (*xclbin_file_name == '\0') {
        snprintf(xclbin_file_name, PATH_MAX, file_patterns[0], *search_dirs, xclbin_name, world.mode, device_name);
    }

    free(device_name);

    return xcl_import_binary_file(world, xclbin_file_name);
}

cl_program xcl_import_source(xcl_world world,
                            const char *krnl_file
) {
	int err;

	char *krnl_bin;
	load_file_to_memory(krnl_file, &krnl_bin);

	cl_program program = clCreateProgramWithSource(world.context, 1,
	                                               (const char**) &krnl_bin,
	                                               0, &err);
	if ((err!=CL_SUCCESS) || (!program))  {
		printf("Error: Failed to create compute program from binary %d!\n",
		       err);
		printf("Test failed\n");
		exit(EXIT_FAILURE);
	}

	err = clBuildProgram(program, 0, NULL, NULL, NULL, NULL);
	if (err != CL_SUCCESS) {
		size_t len;
		char buffer[2048];

		printf("Error: Failed to build program executable!\n");
		clGetProgramBuildInfo(program, world.device_id, CL_PROGRAM_BUILD_LOG,
		                      sizeof(buffer), buffer, &len);
		printf("%s\n", buffer);
		printf("Test failed\n");
		exit(EXIT_FAILURE);
	}

	free(krnl_bin);

	return program;
}

cl_kernel xcl_get_kernel(cl_program program,
                         const char *krnl_name
) {
	int err;

	cl_kernel kernel = clCreateKernel(program, krnl_name, &err);
	if (!kernel || err != CL_SUCCESS) {
		printf("Error: Failed to create kernel for %s: %d\n", krnl_name, err);
		exit(EXIT_FAILURE);
	}

	return kernel;
}

void xcl_set_kernel_arg(cl_kernel krnl,
                        cl_uint num,
                        size_t size,
                        const void *ptr
) {
	int err = clSetKernelArg(krnl, num, size, ptr);

	if(err != CL_SUCCESS) {
		printf("Error: Failed to set kernel arg\n");
		exit(EXIT_FAILURE);
	}
}

cl_mem xcl_malloc(xcl_world world, cl_mem_flags flags, size_t size) {
	cl_mem mem = clCreateBuffer(world.context, flags, size, NULL, NULL);

	if (!mem) {
		printf("Error: Failed to allocate device memory!\n");
		exit(EXIT_FAILURE);
	}

	return mem;
}

void xcl_memcpy_to_device(xcl_world world, cl_mem dest, void* src,
                          size_t size) {
	int err = clEnqueueWriteBuffer(world.command_queue, dest, CL_TRUE, 0, size,
	                               src, 0, NULL, NULL);
	if (err != CL_SUCCESS) {
		printf("Error: Failed to write to source array! %d\n", err);
		exit(EXIT_FAILURE);
	}
}

void xcl_memcpy_from_device(xcl_world world, void* dest,
                            cl_mem src, size_t size
) {
	int err = clEnqueueReadBuffer(world.command_queue, src, CL_TRUE, 0, size,
	                              dest, 0, NULL, NULL);
	if (err != CL_SUCCESS) {
		printf("Error: Failed to read output array! %d\n", err);
		exit(EXIT_FAILURE);
	}
}

unsigned long xcl_get_event_duration(cl_event event) {
	unsigned long start, stop;

	clGetEventProfilingInfo(event, CL_PROFILING_COMMAND_QUEUED,
	                        sizeof(unsigned long), &start, NULL);
	clGetEventProfilingInfo(event, CL_PROFILING_COMMAND_END,
	                        sizeof(unsigned long), &stop, NULL);

	return stop - start;
}

unsigned long xcl_run_kernel3d(xcl_world world, cl_kernel krnl,
                               size_t x, size_t y, size_t z
) {
	size_t size[3] = {x, y, z};
	cl_event event;

	int err = clEnqueueNDRangeKernel(world.command_queue, krnl, 3,
	                                 NULL, size, size, 0, NULL, &event);
	if( err != CL_SUCCESS) {
		printf("Error: failed to execute kernel! %d\n", err);
		exit(EXIT_FAILURE);
	}

	clFinish(world.command_queue);

	return xcl_get_event_duration(event);
}

void xcl_run_kernel3d_nb(xcl_world world, cl_kernel krnl,cl_event *event,
                               size_t x, size_t y, size_t z) {
	size_t size[3] = {x, y, z};

	int err = clEnqueueNDRangeKernel(world.command_queue, krnl, 3,
	                                 NULL, size, size, 0, NULL, event);
	if( err != CL_SUCCESS) {
		printf("Error: failed to execute kernel! %d\n", err);
		printf("Test failed\n");
		exit(EXIT_FAILURE);
	}

	return;
}
