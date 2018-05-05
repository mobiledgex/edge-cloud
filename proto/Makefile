# Makefile

GOPATH		= ../../../..
PROTOCGRPC	= ../../../grpc-ecosystem/grpc-gateway/third_party/googleapis
BUILTIN		= Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types

build:
	protoc -I. -I${GOPATH} -I${PROTOCGRPC} --gomex_out=plugins=grpc+mex,${BUILTIN}:. *.proto
	protoc -I. -I${GOPATH} -I${PROTOCGRPC} --test_out=${BUILTIN}:../testutil *.proto
	#protoc -I. -I${GOPATH} -I${PROTOCGRPC} --gofast_out=plugins=grpc:. *.proto
	#protoc -I. -I${GOPATH} -I${PROTOCGRPC} --go_out=plugins=grpc:. *.proto
