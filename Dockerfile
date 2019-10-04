FROM arm32v7/golang:stretch

COPY qemu-arm-static /usr/bin/
WORKDIR /go/src/github.com/automatedhome/evok-mqtt-bridge
COPY . .
RUN make build

FROM arm32v7/busybox:1.30-glibc

COPY --from=0 /go/src/github.com/automatedhome/evok-mqtt-bridge/evok-mqtt-bridge /usr/bin/evok-mqtt-bridge
COPY --from=0 /go/src/github.com/automatedhome/evok-mqtt-bridge/config.yaml /config.yaml

ENTRYPOINT [ "/usr/bin/evok-mqtt-bridge" ]
