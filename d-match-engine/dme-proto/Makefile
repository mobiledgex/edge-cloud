# Makefile

GOPATH		= ../../../../../..
GW		= ../../vendor/github.com/grpc-ecosystem/grpc-gateway
APIS		= ../../vendor/github.com/gogo/googleapis
IMPORT_ROOT	= ../../../../..
INCLUDE		= -I. -I${GW} -I${APIS} -I${IMPORT_ROOT} -I../../vendor -I${GOPATH}
BUILTIN		= Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types

build:
	protoc ${INCLUDE} --gomex_out=plugins=grpc+mex,${BUILTIN}:. *.proto
