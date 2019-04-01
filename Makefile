# Makefile
include Makedefs

all: build install 

linux: build-linux install-linux

dep:
	dep ensure -update github.com/mobiledgex/edge-cloud-infra
	dep ensure -vendor-only

build:
	make -C protogen
	make -C ./protoc-gen-gomex
	go install ./protoc-gen-test
	go install ./protoc-gen-notify
	go install ./protoc-gen-mc2
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
	for ADDLTAG in ${ADDLTAGS}; do \
		docker tag mobiledgex/edge-cloud:${TAG} $$ADDLTAG; \
		docker push $$ADDLTAG; \
	done

install:
	go install ./...

install-linux:
	${LINUX_XCOMPILE_ENV} go install ./...

tools:
	go install ./vendor/github.com/golang/protobuf/protoc-gen-go
	go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogo
	go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogofast
	go install ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

doc:
	make -C edgeproto doc

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
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/stop_cleanup.yml -setupfile ./setup-env/e2e-tests/setups/local_multi.yml -stop -notimestamp

test-sdk:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/sdk_test/stop_start_create_sdk.yml -setupfile ./setup-env/e2e-tests/setups/local_sdk.yml

# QA testing - manual
test-robot-start:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_start_create_automation.yml -setupfile ./setup-env/e2e-tests/setups/local_multi_automation.yml -stop -notimestamp

test-robot-stop:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/stop_cleanup.yml -setupfile ./setup-env/e2e-tests/setups/local_multi_automation.yml -stop -notimestamp

## note: DIND requires make from edge-cloud-infra to install dependencies
test-dind-start:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/deploy_start_create_dind.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml -notimestamp -stop

test-dind-stop:
	e2e-tests -testfile ./setup-env/e2e-tests/testfiles/delete_dind_stop_cleanup.yml -setupfile ./setup-env/e2e-tests/setups/local_dind.yml -notimestamp


clean:
	go clean ./...
