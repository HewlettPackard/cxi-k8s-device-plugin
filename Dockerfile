FROM docker.io/golang:1.23.1-alpine3.20

RUN apk --no-cache add git pkgconfig build-base libdrm-dev
RUN mkdir -p /go/src/github.com/HPE/cxi-k8s-device-plugin
ADD . /go/src/github.com/HPE/cxi-k8s-device-plugin
WORKDIR /go/src/github.com/HPE/cxi-k8s-device-plugin
RUN go build \
    -ldflags="-X main.version=$(git describe --always --long --dirty)" \
    src/main.go
RUN mv main cxi-k8s-device-plugin
RUN cp cxi-k8s-device-plugin /go/bin/

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /go/bin/cxi-k8s-device-plugin .
CMD ["./cxi-k8s-device-plugin", "-logtostderr=true", "-stderrthreshold=INFO", "-v=5"]