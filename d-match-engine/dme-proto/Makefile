# Makefile

GW		= $(shell go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway)
APIS		= $(shell go list -f '{{ .Dir }}' -m github.com/gogo/googleapis)
INCLUDE		= -I. -I${GW} -I${APIS} -I${GOPATH}
BUILTIN		= Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor
PROTOS		= app-client.proto appcommon.proto dynamic-location-group.proto loc.proto
DMEDIR		= ${GOPATH}/src/github.com/mobiledgex/edge-proto/dme

build: $(PROTOS)
	protoc ${INCLUDE} --gomex_out=plugins=grpc+mex,${BUILTIN}:. *.proto
	protoc ${INCLUDE} --cmd_out=${BUILTIN}:../../gencmd *.proto
	protoc ${INCLUDE} --grpc-gateway_out=${BUILTIN}:. *.proto

# swagger annotations for REST APIs
swagger: $(PROTOS)
	# go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	protoc ${INCLUDE} --swagger_out=logtostderr=true,allow_merge=true,merge_file_name=app-client:. *.proto

external-swagger: $(PROTOS)
	mkdir -p ./external-doc
	protoc ${INCLUDE} --swagger_out=logtostderr=true,allow_merge=true,config_file=swagger_config.yaml,merge_file_name=app-client:./external-doc *.proto


$(PROTOS):
	ln -sf ${DMEDIR}/$@ $@
