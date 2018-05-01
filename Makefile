# Makefile

all: build install

build:
	make -C protogen
	make -C ./protoc-gen-gomex
	make -C proto
	go build ./...

install:
	go install ./...
