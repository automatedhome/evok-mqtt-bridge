FROM golang:1.18 as builder
 
WORKDIR /go/src/github.com/automatedhome/evok-mqtt-bridge
COPY . .
RUN CGO_ENABLED=0 go build -o evok-mqtt-bridge cmd/main.go

FROM busybox:glibc

COPY --from=builder /go/src/github.com/automatedhome/evok-mqtt-bridge/evok-mqtt-bridge /usr/bin/evok-mqtt-bridge
COPY --from=builder /go/src/github.com/automatedhome/evok-mqtt-bridge/config.yaml /config.yaml

ENTRYPOINT [ "/usr/bin/evok-mqtt-bridge" ]
