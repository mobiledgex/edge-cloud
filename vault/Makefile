PKG = "mex-vault"
PROGRAM = mex-vault

.PHONY: build api dep test race msan

default: docker-build docker-push

docker-build:
	./docker-build.sh

docker-push:
	docker tag mobiledgex/mex-vault registry.mobiledgex.net:5000/mobiledgex/mex-vault
	docker push registry.mobiledgex.net:5000/mobiledgex/mex-vault


dep: api ## Fetch dependencies
	@go get ./...


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
