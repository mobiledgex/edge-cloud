# Makefile
all: build install build-linux install-linux

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

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./...
	make -C d-match-engine linux

install:
	go install ./...

install-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install ./...
 
tools:
	go install ./vendor/github.com/golang/protobuf/protoc-gen-go
	go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogo
	go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogofast
	go install ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
