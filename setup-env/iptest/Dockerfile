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
