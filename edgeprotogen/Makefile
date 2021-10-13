# Makefile

APIS		= $(shell go list -f '{{ .Dir }}' -m github.com/gogo/googleapis)
EDGEPROTOGENDIR	= ${GOPATH}/src/github.com/mobiledgex/edge-proto/edgeprotogen
PROTOS		= edgeprotogen.proto

build: $(PROTOS)
	protoc -I. -I${APIS} --gogofast_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. *.proto

$(PROTOS):
	ln -sf ${EDGEPROTOGENDIR}/$@ $@
