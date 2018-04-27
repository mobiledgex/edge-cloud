# Makefile

GOPATH		= ../../../..
PROTOCGRPC	= ../../../grpc-ecosystem/grpc-gateway/third_party/googleapis

build:
	protoc -I. -I${GOPATH} -I${PROTOCGRPC} --gomex_out=plugins=grpc+gomex:. *.proto
	#protoc -I. -I${GOPATH} -I${PROTOCGRPC} --gofast_out=plugins=grpc:. *.proto
	#protoc -I. -I${GOPATH} -I${PROTOCGRPC} --go_out=plugins=grpc:. *.proto
