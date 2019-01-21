TAG ?= latest

build:
	go build ./...

docker-build:
	./docker-build.sh

docker-push:
	docker TAG mobiledgex/metrics-exporter:${TAG} registry.mobiledgex.net:5000/mobiledgex/metrics-exporter:${TAG}
	docker push registry.mobiledgex.net:5000/mobiledgex/metrics-writer:${TAG}

install:
	go install ./...