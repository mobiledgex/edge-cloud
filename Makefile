# Makefile

all: build install

build:
	make -C protogen
	make -C ./protoc-gen-gomex
	go install ./protoc-gen-test
	make -C proto
	make -C d-match-engine
	go build ./...

install:
	go install ./...
