GO111MODULE=on
export GO111MODULE

APP=evok-mqtt-bridge
IMAGE=automatedhome/$(APP)

.PHONY: build
build:
	go build -mod=vendor -o $(APP) cmd/main.go

qemu-arm-static:
	./hooks/post_checkout

.PHONY: image
image: qemu-arm-static
	./hooks/pre_build
	docker build -t $(IMAGE) .
