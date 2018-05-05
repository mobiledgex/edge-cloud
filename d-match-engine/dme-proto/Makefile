# Makefile

GOPATH		= ../../../../../..
PROTO-PATH	= ../../proto
BUILTIN		= Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types

build:
	protoc -I. -I${GOPATH} -I${PROTO-PATH} --go_out=plugins=grpc:. *.proto
