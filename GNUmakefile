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
INTERNAL_CONFIG:=$(KUBECONFIG:$(HOME)%=%)

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

# Register the Kubemove crds in both source and destination cluster
# Example: make register-crds SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: register-crds
register-crds:
	@echo "Registering the CRDs in both source and destination clusters...."
	@docker run                                                          \
			-i                                                           \
			--rm                                                         \
			--net=host                                                   \
			-u $$(id -u):$$(id -g)                                       \
			-v $$(pwd):/go/src/$(MODULE)                                 \
			-v $$(pwd)/.go/cache:/.cache                                 \
			-v $(HOME)/.kube:/.kube                                      \
			-w /go/src/$(MODULE)                                         \
			--env HTTP_PROXY=$(HTTP_PROXY)                               \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                             \
			--env GO111MODULE=on                                         \
			--env GOFLAGS="-mod=vendor"                                  \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                          \
			$(DEV_IMAGE)                                                 \
			/bin/bash -c '                                               \
				kubectl apply -f deploy/crds --context=$(SRC_CONTEXT) && \
				kubectl apply -f deploy/crds --context=$(DST_CONTEXT)    \
			'
	@echo "Done"

# Remove the Kubemove crds from both source and destination cluster
# Example: make remove-crds SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: remove-crds
remove-crds:
	@echo "Registering the CRDs in both source and destination clusters...."
	@docker run                                                           		  \
			-i                                                            		  \
			--rm                                                          		  \
			--net=host                                                    		  \
			-u $$(id -u):$$(id -g)                                        		  \
			-v $$(pwd):/go/src/$(MODULE)                                  		  \
			-v $$(pwd)/.go/cache:/.cache                                  		  \
			-v $(HOME)/.kube:/.kube                                       		  \
			-w /go/src/$(MODULE)                                          		  \
			--env HTTP_PROXY=$(HTTP_PROXY)                                		  \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                              		  \
			--env GO111MODULE=on                                          		  \
			--env GOFLAGS="-mod=vendor"                                   		  \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                           		  \
			$(DEV_IMAGE)                                                  		  \
			/bin/bash -c '                                                		  \
				kubectl delete -f deploy/crds --context=$(SRC_CONTEXT) || true && \
				kubectl delete -f deploy/crds --context=$(DST_CONTEXT)  || true   \
			'
	@echo "Done"

# Create the RBAC resources necessary for Kubemove in both source and destination cluster
# Example: make create-rbac-resources SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: create-rbac-resources
create-rbac-resources:
	@echo "Creating RBAC resources in both source and destination clusters...."
	@docker run                                                               \
			-i                                                                \
			--rm                                                              \
			--net=host                                                        \
			-u $$(id -u):$$(id -g)                                            \
			-v $$(pwd):/go/src/$(MODULE)                                      \
			-v $$(pwd)/.go/cache:/.cache                                      \
			-v $(HOME)/.kube:/.kube                                           \
			-w /go/src/$(MODULE)                                              \
			--env HTTP_PROXY=$(HTTP_PROXY)                                    \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                  \
			--env GO111MODULE=on                                              \
			--env GOFLAGS="-mod=vendor"                                       \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                               \
			$(DEV_IMAGE)                                                      \
			/bin/bash -c '                                                    \
				kubectl apply -f deploy/rbac.yaml --context=$(SRC_CONTEXT) && \
				kubectl apply -f deploy/rbac.yaml --context=$(DST_CONTEXT)    \
			'
	@echo "Done"

# Remove the RBAC resources created for Kubemove from both source and destination cluster
# Example: make create-rbac-resources SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: remove-rbac-resources
remove-rbac-resources:
	@echo "Remove RBAC resources from both source and destination clusters...."
	@docker run                                                                		   \
			-i                                                                 		   \
			--rm                                                               		   \
			--net=host                                                         		   \
			-u $$(id -u):$$(id -g)                                             		   \
			-v $$(pwd):/go/src/$(MODULE)                                       		   \
			-v $$(pwd)/.go/cache:/.cache                                       		   \
			-v $(HOME)/.kube:/.kube                                                	   \
			-w /go/src/$(MODULE)                                               		   \
			--env HTTP_PROXY=$(HTTP_PROXY)                                     		   \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                   		   \
			--env GO111MODULE=on                                               		   \
			--env GOFLAGS="-mod=vendor"                                        		   \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                		   \
			$(DEV_IMAGE)                                                       		   \
			/bin/bash -c '                                                    		   \
				kubectl delete -f deploy/rbac.yaml --context=$(SRC_CONTEXT) || true && \
				kubectl delete -f deploy/rbac.yaml --context=$(DST_CONTEXT) || true    \
			'
	@echo "Done"

# Deploy MovePair controller in both source and destination cluster
# Example: make deploy-mp-ctrl SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: deploy-mp-ctrl
deploy-mp-ctrl:
	@echo "Deploying MovePair controller in both source and destination clusters...."
	@docker run                                                                                                 \
			-i                                                                                                  \
			--rm                                                                                                \
			--net=host                                                                                          \
			-u $$(id -u):$$(id -g)                                                                              \
			-v $$(pwd):/go/src/$(MODULE)                                                                        \
			-v $$(pwd)/.go/cache:/.cache                                                                        \
			-v $(HOME)/.kube:/.kube                                                                             \
			-w /go/src/$(MODULE)                                                                                \
			--env HTTP_PROXY=$(HTTP_PROXY)                                                                      \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                                    \
			--env GO111MODULE=on                                                                                \
			--env GOFLAGS="-mod=vendor"                                                                         \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                                                 \
			--env REGISTRY=$(REGISTRY)                                                                          \
			$(DEV_IMAGE)                                                                                        \
			/bin/bash -c '                                                                                      \
				cat deploy/movepair_controller.yaml | envsubst | kubectl apply -f - --context=$(SRC_CONTEXT) && \
				cat deploy/movepair_controller.yaml | envsubst | kubectl apply -f - --context=$(DST_CONTEXT)    \
			'
	@echo "Done"

# Remove MovePair controller from both source and destination cluster
# Example: make remove-mp-ctrl SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: remove-mp-ctrl
remove-mp-ctrl:
	@echo "Removing MovePair controller from both source and destination clusters...."
	@docker run                                                                               		  \
			-i                                                                                		  \
			--rm                                                                              		  \
			--net=host                                                                        		  \
			-u $$(id -u):$$(id -g)                                                            		  \
			-v $$(pwd):/go/src/$(MODULE)                                                      		  \
			-v $$(pwd)/.go/cache:/.cache                                                      		  \
			-v $(HOME)/.kube:/.kube                                                           		  \
			-w /go/src/$(MODULE)                                                              		  \
			--env HTTP_PROXY=$(HTTP_PROXY)                                                    		  \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                  		  \
			--env GO111MODULE=on                                                              		  \
			--env GOFLAGS="-mod=vendor"                                                       		  \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                               		  \
			--env REGISTRY=$(REGISTRY)                                                        		  \
			$(DEV_IMAGE)                                                                      		  \
			/bin/bash -c '                                                                    		  \
				kubectl delete -f deploy/movepair_controller.yaml --context=$(SRC_CONTEXT) || true && \
				kubectl delete -f deploy/movepair_controller.yaml --context=$(DST_CONTEXT) || true    \
			'
	@echo "Done"

# Deploy MoveEngine controller in both source and destination cluster
# Example: make deploy-me-ctrl SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: deploy-me-ctrl
deploy-me-ctrl:
	@echo "Deploying MoveEngine controller in both source and destination clusters...."
	@docker run                                                                                                   \
			-i                                                                                                    \
			--rm                                                                                                  \
			--net=host                                                                                            \
			-u $$(id -u):$$(id -g)                                                                                \
			-v $$(pwd):/go/src/$(MODULE)                                                                          \
			-v $$(pwd)/.go/cache:/.cache                                                                          \
			-v $(HOME)/.kube:/.kube                                                                               \
			-w /go/src/$(MODULE)                                                                                  \
			--env HTTP_PROXY=$(HTTP_PROXY)                                                                        \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                                      \
			--env GO111MODULE=on                                                                                  \
			--env GOFLAGS="-mod=vendor"                                                                           \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                                                   \
			--env REGISTRY=$(REGISTRY)                                                                            \
			$(DEV_IMAGE)                                                                                          \
			/bin/bash -c '                                                                                        \
				cat deploy/moveengine_controller.yaml | envsubst | kubectl apply -f - --context=$(SRC_CONTEXT) && \
				cat deploy/moveengine_controller.yaml | envsubst | kubectl apply -f - --context=$(DST_CONTEXT)    \
			'
	@echo "Done"

# Remove MoveEngine controller from both source and destination cluster
# Example: make remove-me-ctrl SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: remove-me-ctrl
remove-me-ctrl:
	@echo "Removing MoveEngine controller from both source and destination clusters...."
	@docker run                                                                                 		\
			-i                                                                                  		\
			--rm                                                                                		\
			--net=host                                                                          		\
			-u $$(id -u):$$(id -g)                                                              		\
			-v $$(pwd):/go/src/$(MODULE)                                                        		\
			-v $$(pwd)/.go/cache:/.cache                                                        		\
			-v $(HOME)/.kube:/.kube                                                             		\
			-w /go/src/$(MODULE)                                                                		\
			--env HTTP_PROXY=$(HTTP_PROXY)                                                      		\
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                    		\
			--env GO111MODULE=on                                                                		\
			--env GOFLAGS="-mod=vendor"                                                         		\
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                                 		\
			--env REGISTRY=$(REGISTRY)                                                          		\
			$(DEV_IMAGE)                                                                        		\
			/bin/bash -c '                                                                      		\
				kubectl delete -f deploy/moveengine_controller.yaml --context=$(SRC_CONTEXT) || true && \
				kubectl delete -f deploy/moveengine_controller.yaml --context=$(DST_CONTEXT) || true    \
			'
	@echo "Done"

# Deploy DataSync controller in both source and destination cluster
# Example: make deploy-ds-ctrl SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: deploy-ds-ctrl
deploy-ds-ctrl:
	@echo "Deploying DataSync controller in both source and destination clusters...."
	@docker run                                                                                                 \
			-i                                                                                                  \
			--rm                                                                                                \
			--net=host                                                                                          \
			-u $$(id -u):$$(id -g)                                                                              \
			-v $$(pwd):/go/src/$(MODULE)                                                                        \
			-v $$(pwd)/.go/cache:/.cache                                                                        \
			-v $(HOME)/.kube:/.kube                                                                             \
			-w /go/src/$(MODULE)                                                                                \
			--env HTTP_PROXY=$(HTTP_PROXY)                                                                      \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                                    \
			--env GO111MODULE=on                                                                                \
			--env GOFLAGS="-mod=vendor"                                                                         \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                                                 \
			--env REGISTRY=$(REGISTRY)                                                                          \
			$(DEV_IMAGE)                                                                                        \
			/bin/bash -c '                                                                                      \
				cat deploy/datasync_controller.yaml | envsubst | kubectl apply -f - --context=$(SRC_CONTEXT) && \
				cat deploy/datasync_controller.yaml | envsubst | kubectl apply -f - --context=$(DST_CONTEXT)    \
			'
	@echo "Done"

# Remove DataSync controller from both source and destination cluster
# Example: make remove-ds-ctrl SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: remove-ds-ctrl
remove-ds-ctrl:
	@echo "Removing DataSync controller from both source and destination clusters...."
	@docker run                                                                               		  \
			-i                                                                                		  \
			--rm                                                                              		  \
			--net=host                                                                        		  \
			-u $$(id -u):$$(id -g)                                                            		  \
			-v $$(pwd):/go/src/$(MODULE)                                                      		  \
			-v $$(pwd)/.go/cache:/.cache                                                      		  \
			-v $(HOME)/.kube:/.kube                                                           		  \
			-w /go/src/$(MODULE)                                                              		  \
			--env HTTP_PROXY=$(HTTP_PROXY)                                                    		  \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                  		  \
			--env GO111MODULE=on                                                              		  \
			--env GOFLAGS="-mod=vendor"                                                       		  \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                               		  \
			--env REGISTRY=$(REGISTRY)                                                        		  \
			$(DEV_IMAGE)                                                                      		  \
			/bin/bash -c '                                                                    		  \
				kubectl delete -f deploy/datasync_controller.yaml --context=$(SRC_CONTEXT) || true && \
				kubectl delete -f deploy/datasync_controller.yaml --context=$(DST_CONTEXT) || true    \
			'
	@echo "Done"

# Create a MovePair CR in the source cluster for two local kind clusters
# Example: make create-local-mp SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: create-local-mp
create-local-mp:
	$(eval SRC_CLUSTER_CA:=$(shell (kubectl config view --raw -o=jsonpath='{.clusters[?(@.name=="$(SRC_CONTEXT)")].cluster.certificate-authority-data}')))
	$(eval SRC_CLUSTER_CLIENT_CERT:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(SRC_CONTEXT)")].user.client-certificate-data}')))
	$(eval SRC_CLUSTER_CLIENT_KEY:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(SRC_CONTEXT)")].user.client-key-data}')))
	$(eval DST_CLUSTER_CA:=$(shell (kubectl config view --raw -o=jsonpath='{.clusters[?(@.name=="$(DST_CONTEXT)")].cluster.certificate-authority-data}')))
	$(eval DST_CLUSTER_CLIENT_CERT:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(DST_CONTEXT)")].user.client-certificate-data}')))
	$(eval DST_CLUSTER_CLIENT_KEY:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(DST_CONTEXT)")].user.client-key-data}')))

	@echo "Creating local MovePair CR in the source cluster...."
	@docker run                                                                                                     \
			-i                                                                                                      \
			--rm                                                                                                    \
			--net=host                                                                                              \
			-u $$(id -u):$$(id -g)                                                                                  \
			-v $$(pwd):/go/src/$(MODULE)                                                                            \
			-v $$(pwd)/.go/cache:/.cache                                                                            \
			-v $(HOME)/.kube:/.kube                                                                                 \
			-w /go/src/$(MODULE)                                                                                    \
			--env HTTP_PROXY=$(HTTP_PROXY)                                                                          \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                                        \
			--env GO111MODULE=on                                                                                    \
			--env GOFLAGS="-mod=vendor"                                                                             \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                                                     \
			--env REGISTRY=$(REGISTRY)                                                                              \
			--env SRC_CLUSTER_CA=$(SRC_CLUSTER_CA) 				   												    \
			--env SRC_CLUSTER_CLIENT_CERT=$(SRC_CLUSTER_CLIENT_CERT)												\
			--env SRC_CLUSTER_CLIENT_KEY=$(SRC_CLUSTER_CLIENT_KEY)  												\
			--env SRC_CLUSTER_NAME=$(SRC_CONTEXT)                   												\
			--env SRC_CLUSTER_IP=$(SRC_CLUSTER_IP)                  												\
			--env DST_CLUSTER_CA=$(DST_CLUSTER_CA) 				    												\
			--env DST_CLUSTER_CLIENT_CERT=$(DST_CLUSTER_CLIENT_CERT)												\
			--env DST_CLUSTER_CLIENT_KEY=$(DST_CLUSTER_CLIENT_KEY)  												\
			--env DST_CLUSTER_NAME=$(DST_CONTEXT)                   												\
			--env DST_CLUSTER_IP=$(DST_CLUSTER_IP)                  												\
			$(DEV_IMAGE)                                                                                            \
			/bin/bash -c '                                                                                          \
				cat examples/resources/movepair/local.yaml | envsubst | kubectl apply -f - --context=$(SRC_CONTEXT) \
			'
	@echo "Done"


# Remove local MovePair CR from the source cluster
# Example: make remove-local-mp SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: remove-local-mp
remove-local-mp:
	@echo "Removing local MovePair CR from controller the source cluster...."
	@docker run                                                                                   		  \
			-i                                                                                    		  \
			--rm                                                                                  		  \
			--net=host                                                                            		  \
			-u $$(id -u):$$(id -g)                                                                		  \
			-v $$(pwd):/go/src/$(MODULE)                                                          		  \
			-v $$(pwd)/.go/cache:/.cache                                                          		  \
			-v $(HOME)/.kube:/.kube                                                               		  \
			-w /go/src/$(MODULE)                                                                  		  \
			--env HTTP_PROXY=$(HTTP_PROXY)                                                        		  \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                      		  \
			--env GO111MODULE=on                                                                  		  \
			--env GOFLAGS="-mod=vendor"                                                           		  \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                                   		  \
			--env REGISTRY=$(REGISTRY)                                                            		  \
			$(DEV_IMAGE)                                                                          		  \
			/bin/bash -c '                                                                        		  \
				kubectl delete -f examples/resources/movepair/local.yaml --context=$(SRC_CONTEXT) || true \
			'
	@echo "Done"

# Deploy Minio server in the destination cluster
# Example: make deploy-minio SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: deploy-minio
deploy-minio:
	@echo "Deploying Minio server in the destination cluster...."
	@docker run                                                                                                 				\
			-i                                                                                                  				\
			--rm                                                                                                				\
			--net=host                                                                                          				\
			-u $$(id -u):$$(id -g)                                                                              				\
			-v $$(pwd):/go/src/$(MODULE)                                                                        				\
			-v $$(pwd)/.go/cache:/.cache                                                                        				\
			-v $(HOME)/.kube:/.kube                                                                             				\
			-v /tmp:/.mc																										\
			-w /go/src/$(MODULE)                                                                                				\
			--env HTTP_PROXY=$(HTTP_PROXY)                                                                      				\
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                                    				\
			--env GO111MODULE=on                                                                                				\
			--env GOFLAGS="-mod=vendor"                                                                         				\
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                                                 				\
			--env REGISTRY=$(REGISTRY)                                                                          				\
			--env MINIO_ACCESS_KEY=$(MINIO_ACCESS_KEY) 																			\
			--env MINIO_SECRET_KEY=$(MINIO_SECRET_KEY) 																			\
			$(DEV_IMAGE)                                                                                                        \
			/bin/bash -c '                                                                                                      \
				cat examples/resources/repository/minio_secret.yaml | envsubst | kubectl apply -f - --context=$(SRC_CONTEXT) && \
				cat examples/resources/repository/minio_secret.yaml | envsubst | kubectl apply -f - --context=$(DST_CONTEXT) && \
				cat examples/resources/repository/minio_server.yaml | envsubst | kubectl apply -f - --context=$(DST_CONTEXT)    \
			'
	@echo "Done"

# Remove Minio server from the destination cluster
# Example: make remove-minio SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
.PHONY: remove-minio
remove-minio:
	@echo "Removing Minio server from the destination cluster...."
	@docker run                                                                                               		  \
			-i                                                                                                 		  \
			--rm                                                                                              		  \
			--net=host                                                                                        		  \
			-u $$(id -u):$$(id -g)                                                                            		  \
			-v $$(pwd):/go/src/$(MODULE)                                                                      		  \
			-v $$(pwd)/.go/cache:/.cache                                                                      		  \
			-v $(HOME)/.kube:/.kube                                                                           		  \
			-v /tmp:/.mc																					  		  \
			-w /go/src/$(MODULE)                                                                              		  \
			--env HTTP_PROXY=$(HTTP_PROXY)                                                                    		  \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                                  		  \
			--env GO111MODULE=on                                                                              	  	  \
			--env GOFLAGS="-mod=vendor"                                                                       		  \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                                               		  \
			--env REGISTRY=$(REGISTRY)                                                                        		  \
			$(DEV_IMAGE)                                                                                      		  \
			/bin/bash -c '                                                                                            \
				kubectl delete -f examples/resources/repository/minio_secret.yaml --context=$(SRC_CONTEXT) || true && \
				kubectl delete -f examples/resources/repository/minio_secret.yaml --context=$(DST_CONTEXT) || true && \
				kubectl delete -f examples/resources/repository/minio_server.yaml --context=$(DST_CONTEXT) || true    \
			'
	@echo "Done"

# Create a bucket in the minio server
# Example: make create-minio-bucket BUCKET_NAME=<my-bucket> SRC_CONTEXT=<source context> DST_CONTEXT=<destination context>
BUCKET_NAME:=demo
.PHONY: create-minio-bucket
create-minio-bucket:
	$(eval MINIO_NODEPORT:=$(shell (kubectl get service/minio -o yaml --context=$(DST_CONTEXT) | grep nodePort | cut -c15-)))
	$(eval MINIO_SERVER_ADDRESS:=$(DST_CLUSTER_IP):$(MINIO_NODEPORT))

	@echo "Creating bucket: $(BUCKET_NAME) in the minio server...."
	@docker run                                                                                                       \
			-i                                                                                                  	  \
			--rm                                                                                                	  \
			--net=host                                                                                          	  \
			-u $$(id -u):$$(id -g)                                                                              	  \
			-v $$(pwd):/go/src/$(MODULE)                                                                        	  \
			-v $$(pwd)/.go/cache:/.cache                                                                        	  \
			-v $(HOME)/.kube:/.kube                                                                             	  \
			-v /tmp:/.mc																							  \
			-w /go/src/$(MODULE)                                                                                	  \
			--env HTTP_PROXY=$(HTTP_PROXY)                                                                      	  \
			--env HTTPS_PROXY=$(HTTPS_PROXY)                                                                    	  \
			--env GO111MODULE=on                                                                                	  \
			--env GOFLAGS="-mod=vendor"                                                                         	  \
			--env KUBECONFIG=$(INTERNAL_CONFIG)                                                                 	  \
			--env REGISTRY=$(REGISTRY)                                                                          	  \
			--env MINIO_ACCESS_KEY=$(MINIO_ACCESS_KEY) 																  \
			--env MINIO_SECRET_KEY=$(MINIO_SECRET_KEY) 																  \
			$(DEV_IMAGE)                                                                                              \
			/bin/bash -c '                                                                                            \
				kubectl wait --for=condition=READY pods --all --timeout=5m --context=$(DST_CONTEXT)                && \
				sleep 20 																						   && \
				mc config host add my-minio http://$(MINIO_SERVER_ADDRESS) $(MINIO_ACCESS_KEY) $(MINIO_SECRET_KEY) && \
				mc mb my-minio/$(BUCKET_NAME)  																		  \
			'
	@echo "Done"

# Install necessary Kubemove resources
# Example: make prepare REGISTRY=<your docker registry>
.PHONY: prepare
prepare: register-crds create-rbac-resources deploy-mp-ctrl deploy-me-ctrl deploy-ds-ctrl deploy-minio create-minio-bucket

# Remove all kubemove resources from the source and destination clusters
# Example: make reset
.PHONY: reset
reset: remove-minio remove-ds-ctrl remove-me-ctrl remove-mp-ctrl remove-rbac-resources remove-crds
