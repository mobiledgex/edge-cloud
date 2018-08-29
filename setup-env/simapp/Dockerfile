FROM registry.mobiledgex.net:5000/mobiledgex/build AS build

WORKDIR /go/src/github.com/mobiledgex/edge-cloud/
COPY . .
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV PATH="/go/bin:${PATH}"
WORKDIR /go/src/github.com/mobiledgex/edge-cloud/setup-env/simapp
RUN go get -d -v ./...
RUN go install -v ./...

FROM alpine:latest

COPY --from=build /go/bin/simapp /usr/local/bin

ENTRYPOINT [ "simapp" ]
CMD []
