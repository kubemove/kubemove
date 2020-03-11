SHELL=/bin/bash -o pipefail

PACKAGES = $(shell go list ./... | grep -v 'vendor')

IMG_TAG?=ci
BASE_ENGINE_TAG=ci

REGISTRY?=kubemove
REPO_ROOT:=$(shell pwd)
MODULE=github.com/kubemove/kubemove

GO_VERSION?=1.13.8
GO_PROTO_VERSION?=v1.3.3
OPERATOR_SDK_VERSION?=v0.12.0

DEV_IMG      ?= $(REGISTRY)/kubemove-dev:$(GO_VERSION)
BASE_ENGINE_IMG?=$(REGISTRY)/move-base
ENGINE_IMG?=$(REGISTRY)/move-engine
PAIR_IMG?=$(REGISTRY)/move-pair
DS_IMG?=$(REGISTRY)/move-datasync
PLUGIN_IMG?=$(REGISTRY)/move-plugin

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

.PHONY: bootstrap
bootstrap:
	@go get -u gopkg.in/alecthomas/gometalinter.v1

vet:
	go vet ${PACKAGES}

golint:
	@gometalinter.v1 --install
	@gometalinter.v1 --vendor --disable-all -E errcheck -E misspell ./...

$(BUILD_DIRS):
	@mkdir -p $@

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
	@go build -o $(BIN_DIR)/dummy_plugin cmd/dummy_plugin/main.go
	@echo "Done"

build: datasync engine pair dummy_plugin

clean:
	@echo "Removing old binaries"
	@rm -rf ${BUILD_DIR}
	@echo "Done"

images: engine-image pair-image datasync-image

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
			$(DEV_IMG)                                          \
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
			$(DEV_IMG)                                          \
			/bin/bash -c "operator-sdk generate openapi"
	@echo "Successfully generated crds"

.PHONY: gen-k8s
gen-k8s: $(BUILD_DIRS)
	@echo ""
	@echo "Running operator-sdk generate k8s ...."
	@docker run                                                     \
			-i                                                      \
			--rm                                                    \
			-u $$(id -u):$$(id -g)                                  \
			-v $$(pwd):/src                                         \
			-v $$(pwd)/.go/cache:/.cache                            \
			-w /src                                                 \
			--env HTTP_PROXY=$(HTTP_PROXY)                          \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
			$(DEV_IMG)                                          \
			/bin/bash -c "operator-sdk generate k8s"
	@echo "Done"


# Run gofmt and goimports in all packages except vendor
# Example: make format REGISTRY=<your docker registry>
.PHONY: format
format: $(BUILD_DIRS)
	@echo "Formatting repo...."
	@docker run                                                     			\
			-i                                                      			\
			--rm                                                    			\
			-u $$(id -u):$$(id -g)                                  			\
			-v $$(pwd):/go/src/$(MODULE)                            			\
			-v $$(pwd)/.go/cache:/.cache                            			\
			-w /go/src                                              			\
			--env HTTP_PROXY=$(HTTP_PROXY)                          			\
			--env HTTPS_PROXY=$(HTTPS_PROXY)                        			\
			--env GO111MODULE=on                                    			\
			--env GOFLAGS="-mod=vendor"                             			\
			$(DEV_IMG)                                          			\
			/bin/sh -c "gofmt -s -w $(PACKAGES)	&& goimports -w $(PACKAGES)"
	@echo "Done"

# Run linter
# Example: make lint REGISTRY=<your docker registry>
ADDITIONAL_LINTERS   := goconst,gofmt,goimports,unparam
.PHONY: lint
lint: $(BUILD_DIRS)
	@echo "Running go lint....."
	@docker run                                                     \
			-i                                                      \
			--rm                                                    \
			-u $$(id -u):$$(id -g)                                  \
			-v $$(pwd):/go/src/$(MODULE)                            \
			-v $$(pwd)/.go/cache:/.cache                            \
			-w /go/src/$(MODULE)                                    \
			--env HTTP_PROXY=$(HTTP_PROXY)                          \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
			--env GO111MODULE=on                                    \
			--env GOFLAGS="-mod=vendor"                             \
			$(DEV_IMG)                                          \
			golangci-lint run                                       \
			--enable $(ADDITIONAL_LINTERS)                          \
			--timeout=10m                                           \
			--skip-dirs-use-default                                 \
			--skip-dirs=vendor										\
			--skip-files="zz_generated.*\.go$\"
	@echo "Done"

# Update the dependencies
# Example: make revendor
.PHONY: revendor
revendor:
	@echo "Revendoring project....."
	@docker run                                                     \
			-i                                                      \
			--rm                                                    \
			-u $$(id -u):$$(id -g)                                  \
			-v $$(pwd):/go/src/$(MODULE)                            \
			-v $$(pwd)/.go/cache:/.cache                            \
			-w /go/src/$(MODULE)                                    \
			--env HTTP_PROXY=$(HTTP_PROXY)                          \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
			--env GO111MODULE=on                                    \
			--env GOFLAGS="-mod=vendor"                             \
			$(DEV_IMG)                                          \
			/bin/sh -c "go mod vendor && go mod tidy"
	@echo "Done"


.PHONY: dev-image
dev-image:
	@echo "Building developer image...."
	@docker build -t $(DEV_IMG) -f build/dev.Dockerfile ./build --no-cache \
			--build-arg GO_VERSION=$(GO_VERSION)                                   \
			--build-arg GO_PROTO_VERSION=$(GO_PROTO_VERSION)                       \
			--build-arg OPERATOR_SDK_VERSION=$(OPERATOR_SDK_VERSION)
	@echo "Successfully built developer image"

.PHONY: deploy-images
deploy-images:
	@IMAGE=$(BASE_ENGINE_IMG) ./build/push
	@IMAGE=$(ENGINE_IMG) ./build/push
	@IMAGE=$(PAIR_IMG) ./build/push
	@IMAGE=$(DS_IMG) ./build/push
