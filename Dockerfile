FROM docker.io/golang:1.23.1-alpine3.20
RUN apk --no-cache add git pkgconfig build-base libdrm-dev
RUN mkdir -p /go/src/github.com/HPE/cxi-k8s-device-plugin
ADD . /go/src/HPE/cxi-k8s-device-plugin
WORKDIR /go/src/HPE/cxi-k8s-device-plugin/src
RUN go install \
    -ldflags="-X main.gitDescribe=$(git -C /go/src/HPE/cxi-k8s-device-plugin describe --always --long --dirty)"

FROM alpine:3.20.3
WORKDIR /root/
COPY --from=0 /go/bin/cxi-k8s-device-plugin .
CMD ["./cxi-k8s-device-plugin", "-logtostderr=true", "-stderrthreshold=INFO", "-v=5"]