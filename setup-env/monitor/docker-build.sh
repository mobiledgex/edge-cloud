#!/bin/bash

# this has to be in shell script because we have to run at top level
cd ../..
pwd
cp ~/go/bin/linux_amd64/edgectl setup-env/monitor/edgectl
docker build -t mobiledgex/monitor -f setup-env/monitor/Dockerfile .
rm setup-env/monitor/edgectl
