# Makefile

all: build install

build:
	make -C protogen
	make -C ./protoc-gen-gomex
	go install ./protoc-gen-test
	go install ./protoc-gen-cmd
	make -C edgeproto
	make -C testgen
	make -C d-match-engine
	go build ./...
	go vet ./...

install:
	go install ./...

tools:
	go install ./vendor/github.com/golang/protobuf/protoc-gen-go
	go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogo
	go install ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
