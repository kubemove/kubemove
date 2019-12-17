all: kubemove-cli datasync engine pair

PACKAGES = $(shell go list ./... | grep -v 'vendor')

HUB_USER?=mayankrpatel
ENGINE_IMG?=$(HUB_USER)/move-engine
PAIR_IMG?=$(HUB_USER)/move-pair

IMG_TAG=ci

BASE_ENGINE_IMG?=$(HUB_USER)/move-base
BASE_ENGINE_TAG=ci

BUILD_DIR=build/_output

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

vet:
	go vet ${PACKAGES}

golint:
	@gometalinter --install
	@gometalinter --vendor --disable-all -E errcheck -E misspell ./...

build_dir:
	@mkdir -p ${BUILD_DIR}/bin

kubemove-cli: build_dir
	@echo "Building kubemove-cli"
	@rm -rf _output/bin/kubemove
	@go build -o ${BUILD_DIR}/bin/kubemove cmd/kubemove/main.go
	@echo "Done"

datasync: build_dir
	@echo "Building kubemove-datasync"
	@rm -rf _output/bin/datasync
	@go build -o ${BUILD_DIR}/bin/datasync cmd/datasync/main.go
	@echo "Done"

engine: build_dir
	@echo "Building kubemove-engine"
	@rm -rf _output/bin/kengine
	@go build -o ${BUILD_DIR}/bin/kengine cmd/engine/main.go
	@echo "Done"

pair: build_dir
	@echo "Building kubemove-pair"
	@rm -rf _output/bin/kpair
	@go build -o ${BUILD_DIR}/bin/kpair cmd/pair/main.go
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

base-image:
	@echo "Building base docker image for kubemove"
	@docker build -t $(BASE_ENGINE_IMG):$(BASE_ENGINE_TAG) -f build/Dockerfile-base ./build
