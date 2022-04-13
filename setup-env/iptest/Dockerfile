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

FROM registry.mobiledgex.net:5000/mobiledgex/build AS build

WORKDIR /go/src/github.com/mobiledgex/edge-cloud/
COPY . .
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV PATH="/go/bin:${PATH}"
WORKDIR /go/src/github.com/mobiledgex/edge-cloud/setup-env/iptest/
RUN go get -d -v ./...
RUN go install -v ./...

FROM ubuntu:latest
COPY --from=build /go/bin/iptest /usr/local/bin
ENTRYPOINT [ "iptest" ]
CMD []
