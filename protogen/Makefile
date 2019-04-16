# Makefile

APIS		= $(shell go list -f '{{ .Dir }}' -m github.com/gogo/googleapis)

build:
	protoc -I. -I${APIS} --gogofast_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. *.proto
