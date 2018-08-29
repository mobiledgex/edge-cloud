PKG = "simapp"
GOPATH ?= $(shell go env GOPATH)
GO_PACKAGES := $(shell go list ./... | grep -v /vendor/)
PROGRAM = simapp

.PHONY: build api dep test race msan

default: build docker-build docker-push

docker-build:
	./docker-build.sh

docker-push:
	docker tag mobiledgex/simapp registry.mobiledgex.net:5000/mobiledgex/simapp
	docker push registry.mobiledgex.net:5000/mobiledgex/simapp

build: api dep ## Build simapp
	go get ./...
	mkdir -p build
	CGO_ENABLED=0 go build  -o build/${PROGRAM} .

dep: api ## Fetch dependencies
	@go get ./...


api/simapp.pb.go: api/simapp.proto
	@protoc -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:. \
		api/simapp.proto

api/simapp.pb.gw.go: api/simapp.proto
	@protoc -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:. \
		api/simapp.proto

api/simapp.swagger.json: api/simapp.proto
	@protoc -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--swagger_out=logtostderr=true:. \
		api/simapp.proto

clean: ## Clean compiled binaries
	@rm -f build/${PROGRAM}

realclean: ## Clean compiled binaries and all generated files
	@rm -f build/${PROGRAM}

test: dep ## Run tests
	@go test -short ${GO_PACKAGES}

race: dep ## Run tests with race detector
	@go test -race -short ${GO_PACKAGES}

msan: dep ## Run tests with memory sanitizer
	@go test -msan -short ${GO_PACKAGES}
