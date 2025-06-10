CONTAINER_TOOL ?= docker
USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)
TAG ?= $(shell git describe --dirty --long --always)
PLATFORMS ?= linux/arm64,linux/amd64

.PHONY: build clean verify-tool
.DEFAULT_GOAL: build

build:
	go mod tidy
	go mod verify
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -ldflags "-extldflags '-static'" -o ./dist/discovery-handler ./cmd/secure-http-discovery-handler

clean:
	@rm -rf dist

verify-tool:
	go mod tidy
	go mod verify
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -ldflags "-extldflags '-static'" -o ./dist/dice-verify ./cmd/verify-tool

image:
	$(CONTAINER_TOOL) buildx build --platform $(PLATFORMS) --push -t harbor.nbfc.io/nubificus/secure-http-discovery-handler:$(TAG) -f Dockerfile .