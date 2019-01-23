FROM registry.mobiledgex.net:5000/mobiledgex/build AS build

WORKDIR /go/src/github.com/mobiledgex/edge-cloud/
COPY . .
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV PATH="/go/bin:${PATH}"
WORKDIR /go/src/github.com/mobiledgex/edge-cloud/metrics-exporter
RUN go get -d -v ./...
RUN go build ./...
RUN go install -v ./...

FROM alpine:latest
COPY --from=build /go/bin/metrics-exporter /usr/local/bin

CMD [ "metrics-exporter", "--influxdb", "40.122.54.157:8086" ]
