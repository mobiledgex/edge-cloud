#!/bin/bash
# Copyright 2022 MobiledgeX, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# this has to be in shell script because we have to run at top level
cd ../..
pwd
cp ~/go/bin/linux_amd64/edgectl setup-env/monitor/edgectl
docker build -t mobiledgex/monitor -f setup-env/monitor/Dockerfile .
rm setup-env/monitor/edgectl
