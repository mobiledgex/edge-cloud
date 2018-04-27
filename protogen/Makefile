# Makefile

PROTOCGRPC	= ../../../grpc-ecosystem/grpc-gateway/third_party/googleapis

build:
	protoc -I. -I${PROTOCGRPC} --go_out=. *.proto
