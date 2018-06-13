# Makefile

all: build install

build:
	make -C dme-proto
	go build ./...

linux:
	make -C dme-proto
	GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./...
	GO_ENABLED=0 GOOS=linux GOARCH=amd64 go install ./...

install:
	go install ./...
