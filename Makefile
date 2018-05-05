# Makefile

all: build install

build:
	make -C d-match-engine
	make -C protogen
	make -C ./protoc-gen-gomex
	go install ./protoc-gen-test
	make -C proto
	go build ./...

install:
	go install ./...
