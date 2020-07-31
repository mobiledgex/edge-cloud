TAG ?= $(shell git describe --tags)
IMAGE = docker.mobiledgex.net/mobiledgex/mobiledgex_public/envoy-with-curl:$(TAG)

build:
	docker build -t $(IMAGE) .

publish: build
	docker push $(IMAGE)
