# Makefile

GOPATH		= ../../../..
GW		= $(shell go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway)
APIS		= $(shell go list -f '{{ .Dir }}' -m github.com/gogo/googleapis)
INCLUDE		= -I. -I${GW} -I${APIS} -I${GOPATH}
BUILTIN		= Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types

build:
	protoc ${INCLUDE} --gomex_out=plugins=grpc+mex,${BUILTIN}:. *.proto
	protoc ${INCLUDE} --test_out=${BUILTIN}:../testutil *.proto
	protoc ${INCLUDE} --grpc-gateway_out=${BUILTIN}:. *.proto
	protoc ${INCLUDE} --cmd_out=${BUILTIN}:../gencmd *.proto
	protoc ${INCLUDE} --notify_out=${BUILTIN}:../notify *.proto
	protoc ${INCLUDE} --mc2_out=${BUILTIN}:../mc/orm *.proto
	protoc ${INCLUDE} --mc2_out=${BUILTIN},gentest,suffix=_mc2_test.go:../mc/orm *.proto

doc: build
	protoc ${INCLUDE} --doc_out=./doc --doc_opt=html,index.html *.proto
	protoc ${INCLUDE} --doc_out=./doc --doc_opt=markdown,README.md *.proto
	#swagger annotations for REST APIs
	protoc ${INCLUDE} --swagger_out=logtostderr=true,allow_merge=true:./doc  *.proto
