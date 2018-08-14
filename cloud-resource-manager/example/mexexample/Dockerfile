FROM registry.mobiledgex.net:5000/mobiledgex/build AS build

WORKDIR /go/src/github.com/mobiledgex/edge-cloud/
COPY . .
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV PATH="/go/bin:${PATH}"
WORKDIR /go/src/github.com/mobiledgex/edge-cloud/cloud-resource-manager/example/mexexample
RUN go get -d -v ./...
RUN go install -v ./...

FROM alpine:latest

COPY --from=build /go/bin/mexexample /usr/local/bin

ENTRYPOINT [ "mexexample" ]
CMD []
