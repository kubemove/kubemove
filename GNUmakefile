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

DEV_IMAGE?= $(REGISTRY)/kubemove-dev:$(GO_VERSION)
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

push-images:
	@docker push $(ENGINE_IMG):$(IMG_TAG)
	@docker push $(PAIR_IMG):$(IMG_TAG)
	@docker push $(DS_IMG):$(IMG_TAG)
	@docker push $(PLUGIN_IMG):$(IMG_TAG)
	@docker push $(BASE_ENGINE_IMG):$(BASE_ENGINE_TAG)

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
			$(DEV_IMAGE)                                            \
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
			$(DEV_IMAGE)                                            \
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
			$(DEV_IMAGE)                                            \
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
			$(DEV_IMAGE)                                          			    \
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
			$(DEV_IMAGE)                                            \
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
			$(DEV_IMAGE)                                            \
			/bin/sh -c "go mod vendor && go mod tidy"
	@echo "Done"


.PHONY: dev-image
dev-image:
	@echo "Building developer image...."
	@docker build -t $(DEV_IMAGE) -f build/dev.Dockerfile ./build --no-cache \
			--build-arg GO_VERSION=$(GO_VERSION)                             \
			--build-arg GO_PROTO_VERSION=$(GO_PROTO_VERSION)                 \
			--build-arg OPERATOR_SDK_VERSION=$(OPERATOR_SDK_VERSION)
	@echo "Successfully built developer image"

.PHONY: deploy-images
deploy-images:
	@IMAGE=$(BASE_ENGINE_IMG) ./build/push
	@IMAGE=$(ENGINE_IMG) ./build/push
	@IMAGE=$(PAIR_IMG) ./build/push
	@IMAGE=$(DS_IMG) ./build/push
	@echo ""
	@echo "Pushing developer image...."
	@docker push $(DEV_IMAGE)


KUBECONFIG?=
SRC_CONTEXT?=kind-src-cluster
DST_CONTEXT?=kind-dst-cluster


SRC_ES_NODE_PORT?=
DST_ES_NODE_PORT?=

SRC_CONTROL_PANE:=$(shell /bin/bash -c "kubectl get pod -n kube-system  -o name --context=$(SRC_CONTEXT)| grep kube-apiserver")
SRC_CLUSTER_IP=$(shell (kubectl get -n kube-system $(SRC_CONTROL_PANE) -o yaml --context=$(SRC_CONTEXT)| grep advertise-address= | cut -c27-))
DST_CONTROL_PANE:=$(shell /bin/bash -c "kubectl get pod -n kube-system  -o name --context=$(DST_CONTEXT)| grep kube-apiserver")
DST_CLUSTER_IP=$(shell (kubectl get -n kube-system $(DST_CONTROL_PANE) -o yaml --context=$(DST_CONTEXT)| grep advertise-address= | cut -c27-))

export MINIO_ACCESS_KEY=not@id
export MINIO_SECRET_KEY=not@secret

# Install necessary Kubemove resources
# Example: make prepare REGISTRY=<your docker registry>
.PHONY: prepare
prepare:
	$(eval SRC_CLUSTER_CA:=$(shell (kubectl config view --raw -o=jsonpath='{.clusters[?(@.name=="$(SRC_CONTEXT)")].cluster.certificate-authority-data}')))
	$(eval SRC_CLUSTER_CLIENT_CERT:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(SRC_CONTEXT)")].user.client-certificate-data}')))
	$(eval SRC_CLUSTER_CLIENT_KEY:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(SRC_CONTEXT)")].user.client-key-data}')))
	$(eval DST_CLUSTER_CA:=$(shell (kubectl config view --raw -o=jsonpath='{.clusters[?(@.name=="$(DST_CONTEXT)")].cluster.certificate-authority-data}')))
	$(eval DST_CLUSTER_CLIENT_CERT:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(DST_CONTEXT)")].user.client-certificate-data}')))
	$(eval DST_CLUSTER_CLIENT_KEY:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(DST_CONTEXT)")].user.client-key-data}')))
	@echo "DEV_IMAGE: $(DEV_IMAGE)"
	@docker run                                                     \
			-i                                                      \
			--rm                                                    \
			--net=host                                              \
			-u $$(id -u):$$(id -g)                                  \
			-v $$(pwd):/go/src/$(MODULE)                            \
			-v $$(pwd)/.go/cache:/.cache                            \
			-v $(HOME)/.kube:/.kube                                 \
			-v /tmp:/.mc											\
			-w /go/src/$(MODULE)                                    \
			--env HTTP_PROXY=$(HTTP_PROXY)                          \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
			--env KUBECONFIG=$(KUBECONFIG)                          \
			--env GO111MODULE=on                                    \
			--env GOFLAGS="-mod=vendor"                             \
			--env REGISTRY=$(REGISTRY)                              \
			--env SRC_CLUSTER_CA=$(SRC_CLUSTER_CA) 				    \
			--env SRC_CLUSTER_CLIENT_CERT=$(SRC_CLUSTER_CLIENT_CERT)\
			--env SRC_CLUSTER_CLIENT_KEY=$(SRC_CLUSTER_CLIENT_KEY)  \
			--env SRC_CLUSTER_NAME=$(SRC_CONTEXT)                   \
			--env SRC_CLUSTER_IP=$(SRC_CLUSTER_IP)                  \
			--env DST_CLUSTER_CA=$(DST_CLUSTER_CA) 				    \
			--env DST_CLUSTER_CLIENT_CERT=$(DST_CLUSTER_CLIENT_CERT)\
			--env DST_CLUSTER_CLIENT_KEY=$(DST_CLUSTER_CLIENT_KEY)  \
			--env DST_CLUSTER_NAME=$(DST_CONTEXT)                   \
			--env DST_CLUSTER_IP=$(DST_CLUSTER_IP)                  \
			--env MINIO_ACCESS_KEY=$(MINIO_ACCESS_KEY) 				\
            --env MINIO_SECRET_KEY=$(MINIO_SECRET_KEY) 				\
			$(DEV_IMAGE)                                            \
			/bin/bash -c "                                          \
				KUBECONFIG=$${KUBECONFIG#$(HOME)}                   \
				./hack/prepare.sh                                   \
				"
	@echo "Done"

# Remove all kubemove resources from the source and destination clusters
# Example: make reset
.PHONY: reset
reset:
	@echo "Removing all Kubemove resources from the source and destination clusters...."
	@docker run                                                     \
			-i                                                      \
			--rm                                                    \
			--net=host                                              \
			-u $$(id -u):$$(id -g)                                  \
			-v $$(pwd):/go/src/$(MODULE)                            \
			-v $$(pwd)/.go/cache:/.cache                            \
			-v $(HOME)/.kube:/.kube                                 \
			-v /tmp:/.mc											\
			-w /go/src/$(MODULE)                                    \
			--env HTTP_PROXY=$(HTTP_PROXY)                          \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
			--env KUBECONFIG=$(KUBECONFIG)                          \
			--env GO111MODULE=on                                    \
			--env GOFLAGS="-mod=vendor"                             \
			--env REGISTRY=$(REGISTRY)                              \
			--env SRC_CLUSTER_NAME=$(SRC_CONTEXT)                   \
			--env DST_CLUSTER_NAME=$(DST_CONTEXT)                   \
			$(DEV_IMAGE)                                            \
			/bin/bash -c "                                          \
				KUBECONFIG=$${KUBECONFIG#$(HOME)}                   \
				./hack/reset.sh                                   \
				"
	@echo "Done"

# Create MoveEngine CR to sync data between two Elasticsearch
# Example: make stup-es-sync
.PHONY: setup-es-sync
setup-es-sync:
	@echo " "
	$(eval MINIO_NODEPORT:=$(shell (kubectl get service minio -o yaml --context=$(DST_CONTEXT) | grep nodePort | cut -c15-)))
	$(eval MINIO_SERVER_ADDRESS:=$(DST_CLUSTER_IP):$(MINIO_NODEPORT))
	@echo "Creating MoveEngine CR into the source cluster...."
	@cat examples/resources/moveengine/es-moveengine.yaml |          \
		MINIO_SERVER_ADDRESS=$(MINIO_SERVER_ADDRESS)                 \
		envsubst | kubectl apply -f - --context=$(SRC_CONTEXT)
	@echo "Done"
# Delete MoveEngine CR
# Example: make remove-es-sync
.PHONY: remove-es-sync
remove-es-sync:
	@kubectl delete -f examples/resources/moveengine/es-moveengine.yaml --context=$(SRC_CONTEXT) || true
	@kubectl delete -f examples/resources/moveengine/es-moveengine.yaml --context=$(DST_CONTEXT) || true
