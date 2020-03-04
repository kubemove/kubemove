SHELL=/bin/bash -o pipefail

PACKAGES = $(shell go list ./... | grep -v 'vendor')

HUB_USER?=mayankrpatel
ENGINE_IMG?=$(HUB_USER)/move-engine
PAIR_IMG?=$(HUB_USER)/move-pair
DS_IMG?=$(HUB_USER)/move-ds
PLUGIN_IMG?=$(HUB_USER)/move-plugin

IMG_TAG?=ci

BASE_ENGINE_IMG?=$(HUB_USER)/move-base
BASE_ENGINE_TAG=ci

DOCKER_REGISTRY?=kubemovedev
REPO_ROOT:=$(shell pwd)

GO_VERSION?=1.13.8
GO_PROTO_VERSION?=v1.3.3
OPERATOR_SDK_VERSION?=v0.12.0

BUILD_IMAGE      ?= $(DOCKER_REGISTRY)/kubemove-dev:$(GO_VERSION)

# directories required to build the project
BUILD_DIRS:= build/_output \
			.go/cache
# directory where the go build output will go
BIN_DIR:=build/_output/bin

# Use IMAGE_TAG to <branch-name>-<commit-hash> to make sure that
# one dev don't overwrite another's image in the docker registry
git_branch       := $(shell git rev-parse --abbrev-ref HEAD)
git_tag          := $(shell git describe --exact-match --abbrev=0 2>/dev/null || echo "")
commit_hash      := $(shell git rev-parse --verify --short HEAD)

ifdef git_tag
	IMAGE_TAG = $(git_tag)
else
	IMAGE_TAG = $(git_branch)-$(commit_hash)
endif

all: datasync engine pair dummy_plugin

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

vet:
	go vet ${PACKAGES}

golint:
	@gometalinter --install
	@gometalinter --vendor --disable-all -E errcheck -E misspell ./...

$(BUILD_DIRS):
	@mkdir -p $@

kubemove-cli: $(BUILD_DIRS)
	@echo "Building kubemove-cli"
	@rm -rf _output/bin/kubemove
	@go build -o $(BIN_DIR)/kubemove cmd/kubemove/main.go
	@echo "Done"

datasync: $(BUILD_DIRS)
	@echo "Building kubemove-datasync"
	@rm -rf _output/bin/datasync
	@go build -o $(BIN_DIR)/datasync cmd/datasync/main.go
	@echo "Done"

engine: $(BUILD_DIRS)
	@echo "Building kubemove-engine"
	@rm -rf _output/bin/kengine
	@go build -o $(BIN_DIR)/kengine cmd/engine/main.go
	@echo "Done"

pair: $(BUILD_DIRS)
	@echo "Building kubemove-pair"
	@rm -rf _output/bin/kpair
	@go build -o $(BIN_DIR)/kpair cmd/pair/main.go
	@echo "Done"

dummy_plugin: $(BUILD_DIRS)
	@echo "Building dummy plugin"
	@rm -rf _output/bin/dummy_plugin
	@go build -o $(BIN_DIR)/dummy_plugin cmd/dummy_plugin/dummy_plugin.go
	@echo "Done"

clean:
	@echo "Removing old binaries"
	@rm -rf ${BUILD_DIR}
	@echo "Done"

engine-image: base-image engine
	@echo "Building docker image for kubemove-engine"
	@docker build -t $(ENGINE_IMG):$(IMG_TAG) -f build/Dockerfile-engine --build-arg BASE_IMG=${BASE_ENGINE_IMG}:${BASE_ENGINE_TAG} ./build

pair-image: base-image pair
	@echo "Building docker image for kubemove-pair"
	@docker build -t $(PAIR_IMG):$(IMG_TAG) -f build/Dockerfile-pair --build-arg BASE_IMG=${BASE_ENGINE_IMG}:${BASE_ENGINE_TAG}  ./build

datasync-image: datasync
	@echo "Building docker image for kubemove-datasync"
	@docker build -t $(DS_IMG):$(IMG_TAG) -f build/Dockerfile-ds ./build

dummy_plugin-image: dummy_plugin
	@echo "Building docker image for dummy kubemove-plugin"
	@docker build -t $(PLUGIN_IMG):$(IMG_TAG) -f build/Dockerfile-plugin ./build


base-image:
	@echo "Building base docker image for kubemove"
	@docker build -t $(BASE_ENGINE_IMG):$(BASE_ENGINE_TAG) -f build/Dockerfile-base ./build

.PHONY: gen
gen: gen-crds gen-k8s gen-grpc

.PHONY: gen-grpc
gen-grpc:
	@echo ""
	@echo "Generating gRPC server and client from .proto file...."
	@docker run                                                     \
			-i                                                      \
			--rm                                                    \
			-u $$(id -u):$$(id -g)                                  \
			-v $$(pwd):/src                                         \
			-w /src                                                 \
			--env HTTP_PROXY=$(HTTP_PROXY)                          \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
			$(BUILD_IMAGE)                                          \
			/bin/bash -c "protoc --go_out=plugins=grpc:. pkg/plugin/proto/*.proto"
	@echo "Successfully generated gRPC codes"

.PHONY: gen-crds
gen-crds: $(BUILD_DIRS)
	@echo ""
	@echo "Generating crds using operator sdk...."
	@docker run                                                     \
			-i                                                      \
			--rm                                                    \
			-u $$(id -u):$$(id -g)                                  \
			-v $$(pwd):/src                                         \
			-v $$(pwd)/.go/cache:/.cache                            \
			-w /src                                                 \
			--env HTTP_PROXY=$(HTTP_PROXY)                          \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
			$(BUILD_IMAGE)                                          \
			/bin/bash -c "operator-sdk generate openapi"
	@echo "Successfully generated crds"

.PHONY: dev-image
dev-image:
	@echo "Building developer image...."
	@docker build -t $(BUILD_IMAGE) -f build/dev.Dockerfile ./build --no-cache \
			--build-arg GO_VERSION=$(GO_VERSION)                                   \
			--build-arg GO_PROTO_VERSION=$(GO_PROTO_VERSION)                       \
			--build-arg OPERATOR_SDK_VERSION=$(OPERATOR_SDK_VERSION)
	@echo "Successfully built developer image"
	@echo ""
	@echo "Pushing developer image...."
	@docker push $(BUILD_IMAGE)
