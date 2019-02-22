PKG = "iptest"
GOPATH ?= $(shell go env GOPATH)
GO_PACKAGES := $(shell go list ./... | grep -v /vendor/)
PROGRAM = iptest

.PHONY: build api dep test race msan

default: build docker-build docker-push

docker-build:
	./docker-build.sh

docker-push:
	docker tag mobiledgex/iptest registry.mobiledgex.net:5000/mobiledgex/iptest
	docker push registry.mobiledgex.net:5000/mobiledgex/iptest

