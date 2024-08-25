IMAGE_NAME := "chadgeary/cert-manager-webhook-duckdns"
IMAGE_TAG := "1.0.0"

OUT := $(shell pwd)/_out

$(shell mkdir -p "$(OUT)")

build:
	docker buildx build --platform linux/arm64/v8,linux/amd64 -t "$(IMAGE_NAME):$(IMAGE_TAG)" --push  .
