# Makefile

GOPATH		= ../../../..
GOGO		= $(shell go list -f '{{ .Dir }}' -m github.com/gogo/protobuf)
INCLUDE		= -I. -I${GOGO} -I${GOPATH}
BUILTIN		= Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types,Mgogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto

build:
	protoc ${INCLUDE} --gomex_out=plugins=grpc+mex,${BUILTIN}:. *.proto
	protoc ${INCLUDE} --test_out=${BUILTIN}:./testutil *.proto
	protoc ${INCLUDE} --grpc-gateway_out=${BUILTIN}:. *.proto
	protoc ${INCLUDE} --cmd_out=${BUILTIN}:./gencmd *.proto
