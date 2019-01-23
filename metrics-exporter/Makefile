TAG ?= latest

build:
	go build ./...

build-docker:
	./docker-build.sh
	docker tag mobiledgex/metrics-exporter:${TAG} registry.mobiledgex.net:5000/mobiledgex/metrics-exporter:${TAG}
	docker push registry.mobiledgex.net:5000/mobiledgex/metrics-exporter:${TAG}

install:
	go install ./...