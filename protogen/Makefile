# Makefile

PROTOCGRPC	= ../../../grpc-ecosystem/grpc-gateway/third_party/googleapis
GOGOPROTO	= 

build:
	go install github.com/gogo/protobuf/protoc-gen-gogo
	protoc -I. -I${PROTOCGRPC} --gogofast_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. *.proto
