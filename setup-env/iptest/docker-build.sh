#!/bin/bash
# this has to be in shell script because we have to run at top level
cd ../..
pwd
docker build -t mobiledgex/iptest -f setup-env/iptest/Dockerfile .

