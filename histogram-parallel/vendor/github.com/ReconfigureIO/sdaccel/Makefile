# variable definitions
NAME := sdaccel
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
PKG_RELEASE ?= 1
PROJECT_URL := "https://github.com/ReconfigureIO/$(NAME)"

.PHONY: test all clean compile

CMD_SOURCES := $(shell go list ./... | grep /cmd/)
TARGETS := $(patsubst github.com/ReconfigureIO/sdaccel/cmd/%,dist/%,$(CMD_SOURCES))

all: ${TARGETS}

test:
	go test -v $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)

compile:
	LIBRARY_PATH=${XILINX_SDX}/runtime/lib/x86_64/:${XILINX_SDX}/SDK/lib/lnx64.o/:/usr/lib/x86_64-linux-gnu:${LIBRARY_PATH} CGO_CFLAGS=-I${XILINX_SDX}/runtime/include/1_2/ go build -tags opencl github.com/ReconfigureIO/sdaccel/xcl

dist:
	mkdir -p dist

dist/%: cmd/% | dist
	go build -ldflags "$(LDFLAGS)" -o $@ github.com/ReconfigureIO/sdaccel/$<

clean:
	rm -rf dist
