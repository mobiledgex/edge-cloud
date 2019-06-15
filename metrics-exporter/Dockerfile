FROM golang AS BUILD

WORKDIR /go/src/github.com/mobiledgex/edge-cloud/
COPY . .
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV PATH="/go/bin:${PATH}"
ENV GO111MODULE=on
WORKDIR /go/src/github.com/mobiledgex/edge-cloud/metrics-exporter
RUN go build ./...
RUN go install -v ./...

FROM alpine:latest
COPY --from=build /go/bin/metrics-exporter /usr/local/bin

CMD [ "metrics-exporter", "--influxdb", "https://mexdemo.influxdb.mobiledgex.net:8086" ]
