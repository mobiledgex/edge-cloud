PKG = "monitor"
PROGRAM = monitor

default: build docker-build docker-push

docker-build:
	./docker-build.sh

docker-push:
	docker tag mobiledgex/monitor registry.mobiledgex.net:5000/mobiledgex/monitor
	docker push registry.mobiledgex.net:5000/mobiledgex/monitor

