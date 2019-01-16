# Makefile
include Makedefs

all: build install 

linux: build-linux install-linux

build:
	make -C protogen
	make -C ./protoc-gen-gomex
	go install ./protoc-gen-test
	go install ./protoc-gen-notify
	make -C ./protoc-gen-cmd
	make -C ./log
	make -C edgeproto
	make -C testgen
	make -C d-match-engine
	go build ./...
	go vet ./...

build-linux:
	${LINUX_XCOMPILE_ENV} go build ./...
	make -C d-match-engine linux

build-docker:
	docker build -t mobiledgex/edge-cloud:${TAG} -f docker/Dockerfile.edge-cloud .
	docker tag mobiledgex/edge-cloud:${TAG} registry.mobiledgex.net:5000/mobiledgex/edge-cloud:${TAG}
	docker push registry.mobiledgex.net:5000/mobiledgex/edge-cloud:${TAG}

install:
	go install ./...

install-linux:
	${LINUX_XCOMPILE_ENV} go install ./...

tools:
	go install ./vendor/github.com/golang/protobuf/protoc-gen-go
	go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogo
	go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogofast
	go install ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

test:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/regression_group.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml

test-debug:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/regression_group.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml -stop -notimestamp

# will 1)export PYTHONPATH 2)stop and start processes 3)run python testscases
# can use: "make test-python" to run all tests
# can use: "make test-python dir=controller/operator" to run only the operator testcases or any directory specified
test-python:
	bash ./setup-env/e2e-tests/python/tools/run_all_tests.sh

# start/restart local processes to run individual python or other tests against
test-start:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_start_create.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml -stop -notimestamp

# restart process, clean data
test-reset:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_reset_create.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml -stop -notimestamp

test-stop:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/stop_cleanup.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml -stop -notimestamp

test-sdk:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/sdk_test/stop_start_create_sdk.yml -setupfile ./setup-env/e2e-tests/setups/local_sdk.yml

## note: DIND requires make from edge-cloud-infra to install dependencies
test-dind-start:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_start_create_dind.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml

test-dind-stop:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/delete_dind_stop_cleanup.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml


clean:
	go clean ./...
