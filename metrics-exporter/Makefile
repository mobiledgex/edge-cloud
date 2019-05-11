TAG ?= latest

build:
	go build ./...

build-docker:
	./docker-build.sh
	docker tag mobiledgex/metrics-exporter:${TAG} docker.mobiledgex.net/mobiledgex/images/metrics-exporter:${TAG}
	docker push docker.mobiledgex.net/mobiledgex/images/metrics-exporter:${TAG}

install:
	go install ./...
