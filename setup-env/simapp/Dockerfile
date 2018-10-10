FROM registry.mobiledgex.net:5000/mobiledgex/build AS build

WORKDIR /go/src/github.com/mobiledgex/edge-cloud/
COPY . .
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV PATH="/go/bin:${PATH}"
WORKDIR /go/src/github.com/mobiledgex/edge-cloud/setup-env/simapp/
RUN go get -d -v ./...
RUN go install -v ./...

FROM alpine:latest
COPY --from=build /go/bin/simapp /usr/local/bin
COPY --from=build /go/src/github.com/mobiledgex/edge-cloud/setup-env/simapp/download-files/download_1MB.txt /root/downloadfiles/download_1MB.txt
COPY --from=build /go/src/github.com/mobiledgex/edge-cloud/setup-env/simapp/download-files/download_5MB.txt /root/downloadfiles/download_5MB.txt
COPY --from=build /go/src/github.com/mobiledgex/edge-cloud/setup-env/simapp/download-files/download_10MB.txt /root/downloadfiles/download_10MB.txt
COPY --from=build /go/src/github.com/mobiledgex/edge-cloud/setup-env/simapp/download-files/download_20MB.txt /root/downloadfiles/download_20MB.txt

ENTRYPOINT [ "simapp" ]
CMD []
