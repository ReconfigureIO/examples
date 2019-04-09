# variable definitions
NAME := fixed
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
PKG_RELEASE ?= 1
PROJECT_URL := "https://github.com/ReconfigureIO/$(NAME)"

.PHONY: test vendor install

test:
	go build fixed.go
	go test github.com/ReconfigureIO/fixed/host

vendor: examples/mult/vendor/github.com/ReconfigureIO/$(NAME)/fixed.go

install: vendor
	cd examples/mult && glide install

examples/mult/vendor/github.com/ReconfigureIO/$(NAME)/fixed.go: fixed.go
	mkdir -p examples/mult/vendor/github.com/ReconfigureIO/$(NAME)
	cp -R fixed.go host examples/mult/vendor/github.com/ReconfigureIO/$(NAME)
