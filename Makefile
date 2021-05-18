# Makefile
include Makedefs

GOVERS = $(shell go version | awk '{print $$3}' | cut -d. -f1,2)

export GO111MODULE=on

all: build install 

linux: build-linux install-linux

build-vers:
	(cd version; ./version.sh)

check-vers: build-vers
	@if test $(GOVERS) != go1.15; then \
		echo "Go version is $(GOVERS)"; \
		echo "See https://mobiledgex.atlassian.net/wiki/spaces/SWDEV/pages/307986555/Upgrade+to+go+1.12"; \
		exit 2; \
	fi


build: check-vers
	make -f Makefile.tools
	make -C protogen
	make -C ./protoc-gen-gomex
	go install ./protoc-gen-test
	go install ./protoc-gen-notify
	make -C ./protoc-gen-cmd
	make -C ./log
	make -C d-match-engine/dme-proto
	make -C edgeproto
	make -C testgen
	make -C d-match-engine
	go build ./...
	go vet ./...

build-linux:
	${LINUX_XCOMPILE_ENV} go build ./...
	make -C d-match-engine linux

build-docker:
	rsync --checksum .dockerignore ../.dockerignore
	docker build --build-arg BUILD_TAG="$(shell git describe --always --dirty=+), $(shell date +'%Y-%m-%d'), ${TAG}" \
		--build-arg REGISTRY=$(REGISTRY) \
		-t mobiledgex/edge-cloud:$(TAG) -f docker/Dockerfile.edge-cloud ..
	docker tag mobiledgex/edge-cloud:$(TAG) $(REGISTRY)/edge-cloud:${TAG}
	docker push $(REGISTRY)/edge-cloud:$(TAG)
	docker tag mobiledgex/edge-cloud:$(TAG) $(REGISTRY)/edge-cloud:latest
	docker push $(REGISTRY)/edge-cloud:latest

build-nightly: REGISTRY = harbor.mobiledgex.net/mobiledgex
build-nightly: build-docker
	docker tag mobiledgex/edge-cloud:$(TAG) $(REGISTRY)/edge-cloud:nightly
	docker push $(REGISTRY)/edge-cloud:nightly

install:
	go install ./...

install-linux:
	${LINUX_XCOMPILE_ENV} go install ./...

GOGOPROTO	= $(shell GO111MODULE=on go list -f '{{ .Dir }}' -m github.com/gogo/protobuf)
GRPCGATEWAY	= $(shell GO111MODULE=on go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway)

tools:
	make -f Makefile.tools

doc:
	make -C edgeproto doc

external-doc:
	make -C edgeproto external-doc

lint:
	(cd $(GOPATH)/src/github.com/uber/prototool; go install ./cmd/prototool)
	$(RM) link-gogo-protobuf
	$(RM) link-grpc-gateway
	ln -s $(GOGOPROTO) link-gogo-protobuf
	ln -s $(GRPCGATEWAY) link-grpc-gateway
	prototool lint edgeproto
	prototool lint d-match-engine

UNIT_TEST_LOG ?= /tmp/edge-cloud-unit-test.log

unit-test:
	go test ./... > $(UNIT_TEST_LOG) || !(grep -A6 "\--- FAIL:" $(UNIT_TEST_LOG) && grep "FAIL\tgithub.com" $(UNIT_TEST_LOG))

test:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/regression_group.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml

test-debug:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/regression_group.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml -stop -notimestamp

# start/restart local processes to run individual python or other tests against
test-start:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_start_create.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml -stop -notimestamp

# restart process, clean data
test-reset:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_reset_create.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml -stop -notimestamp

test-stop:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/delete_stop_create.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml -notimestamp

test-sdk:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/sdk_test/stop_start_create_sdk.yml -setupfile ./setup-env/e2e-tests/setups/local_sdk.yml

## note: DIND requires make install-dind to be run once
install-dind:
	./install-dind.sh

test-dind-start:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_start_create_dind.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml -notimestamp -stop

test-dind-stop:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/delete_dind_stop_cleanup.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml -notimestamp

test-dind-docker-start:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_start_create_dind_docker.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml -notimestamp -stop

test-dind-docker-stop:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/delete_dind_stop_cleanup_docker.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml -notimestamp

# requires kind to be installed
test-kind-start:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_start_create_kind.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml -notimestamp -stop

test-kind-stop:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/delete_kind_stop_cleanup.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml -notimestamp



clean: build-vers
	go clean ./...
