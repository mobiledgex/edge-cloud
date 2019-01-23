#!/bin/bash
# this has to be in shell script because we have to run at top level
cd ..
pwd
TAG=latest
docker build -t mobiledgex/metrics-exporter:${TAG} -f metrics-exporter/Dockerfile .

