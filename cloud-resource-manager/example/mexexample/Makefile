PKG = "mexexample"
GOPATH ?= $(shell go env GOPATH)
GO_PACKAGES := $(shell go list ./... | grep -v /vendor/)
PROGRAM = mexexample
API_FILES = api/mexexample.pb.go api/mexexample.pb.gw.go api/mexexample.swagger.json

.PHONY: build api dep test race msan

default: build docker-build docker-push

docker-build:
	docker build -t mobiledgex/mexexample .

# docker login first!
docker-push:
	docker tag mobiledgex/mexexample registry.mobiledgex.net:5000/mobiledgex/mexexample
	docker push registry.mobiledgex.net:5000/mobiledgex/mexexample

build: api dep ## Build mexexample
	go get ./...
	@mkdir -p build
	@CGO_ENABLED=0 go build  -o build/${PROGRAM} .

dep: api ## Fetch dependencies
	@go get ./...

api: $(API_FILES) ## Auto-generate gRPC/REST Go sources

api/mexexample.pb.go: api/mexexample.proto
	@protoc -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:. \
		api/mexexample.proto

api/mexexample.pb.gw.go: api/mexexample.proto
	@protoc -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:. \
		api/mexexample.proto

api/mexexample.swagger.json: api/mexexample.proto
	@protoc -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--swagger_out=logtostderr=true:. \
		api/mexexample.proto

clean: ## Clean compiled binaries
	@rm -f build/${PROGRAM}

realclean: ## Clean compiled binaries and all generated files
	@rm -f build/${PROGRAM}
	@rm -f ${API_FILES}

test: dep ## Run tests
	@go test -short ${GO_PACKAGES}

race: dep ## Run tests with race detector
	@go test -race -short ${GO_PACKAGES}

msan: dep ## Run tests with memory sanitizer
	@go test -msan -short ${GO_PACKAGES}
