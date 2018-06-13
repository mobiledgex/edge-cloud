# Makefile

include ../Makedefs

all: build install

build:
	make -C dme-proto
	go build ./...

linux:
	make -C dme-proto
	${LINUX_XCOMPILE_ENV} go build ./...
	${LINUX_XCOMPILE_ENV} go install ./...

install:
	go install ./...
