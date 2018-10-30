# Makefile

GOPATH		= ../../../..
GW		= ../vendor/github.com/grpc-ecosystem/grpc-gateway
APIS		= ../vendor/github.com/gogo/googleapis
#APIS		= ../../../grpc-ecosystem/grpc-gateway/third_party/googleapis
INCLUDE		= -I. -I${GW} -I${APIS} -I../vendor -I${GOPATH}
BUILTIN		= Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types

build:
	protoc ${INCLUDE} --gomex_out=plugins=grpc+mex,${BUILTIN}:. *.proto
	protoc ${INCLUDE} --test_out=${BUILTIN}:../testutil *.proto
	protoc ${INCLUDE} --grpc-gateway_out=${BUILTIN}:. *.proto
	protoc ${INCLUDE} --cmd_out=${BUILTIN}:../gencmd *.proto

	# to use protodoc, make sure you run:
	# go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc
	# go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	protoc ${INCLUDE} --doc_out=./doc --doc_opt=html,index.html *.proto
	protoc ${INCLUDE} --doc_out=./doc --doc_opt=markdown,README.md *.proto
	#swagger annotations for REST APIs
	protoc ${INCLUDE} --swagger_out=logtostderr=true,allow_merge=true:./doc  *.proto
