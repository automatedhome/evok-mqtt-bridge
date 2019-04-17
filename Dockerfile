FROM arm32v7/golang:stretch

COPY qemu-arm-static /usr/bin/
WORKDIR /go/src/github.com/automatedhome/evok-mqtt-bridge
COPY . .
RUN go build -o bridge cmd/main.go

FROM arm32v7/busybox:1.30-glibc

COPY --from=0 /go/src/github.com/automatedhome/evok-mqtt-bridge/bridge /usr/bin/bridge

ENTRYPOINT [ "/usr/bin/bridge" ]
