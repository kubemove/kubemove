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

all: datasync engine pair dummy_plugin

.PHONY: bootstrap
bootstrap:
	@go get -u gopkg.in/alecthomas/gometalinter.v1

vet:
	go vet ${PACKAGES}

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
ADDITIONAL_LINTERS   := goconst,gofmt,goimports,unparam,misspell
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

KUBECONFIG?=

SRC_CONTEXT?=kind-src-cluster
DST_CONTEXT?=kind-dst-cluster

SRC_ES_NODE_PORT?=
DST_ES_NODE_PORT?=

# Register the Kubemove crds in both source and destination cluster
# Example: make register_crds
.PHONY: register_crds
register_crds:
	@echo "Registering the CRDs in both source and destination clusters...."
	@kubectl apply -f deploy/crds --context=$(SRC_CONTEXT)
	kubectl apply -f deploy/crds --context=$(DST_CONTEXT)
	@echo "Done"

# Remove the Kubemove crds from both source and destination cluster
# Example: make remove_crds
.PHONY: remove_crds
remove_crds:
	@echo "Registering the CRDs in both source and destination clusters...."
	@kubectl delete -f deploy/crds --context=$(SRC_CONTEXT)
	@kubectl delete -f deploy/crds --context=$(DST_CONTEXT)
	@echo "Done"

# Create the RBAC resources necessary for Kubemove in both source and destination cluster
# Example: make create_rbac_resources
.PHONY: create_rbac_resources
create_rbac_resources:
	@echo "Creating RBAC resources in both source and destination clusters...."
	@kubectl apply -f deploy/rbac.yaml --context=$(SRC_CONTEXT)
	@kubectl apply -f deploy/rbac.yaml --context=$(DST_CONTEXT)
	@echo "Done"

# Remove the RBAC resources created for Kubemove from both source and destination cluster
# Example: make create_rbac_resources
.PHONY: remove_rbac_resources
remove_rbac_resources:
	@echo "Remove RBAC resources from both source and destination clusters...."
	@kubectl delete -f deploy/rbac.yaml --context=$(SRC_CONTEXT) || true
	@kubectl delete -f deploy/rbac.yaml --context=$(DST_CONTEXT) || true
	@echo "Done"

# Deploy MovePair controller in both source and destination cluster
# Example: make deploy_mp_ctrl
.PHONY: deploy_mp_ctrl
deploy_mp_ctrl:
	@echo "Deploying MovePair controller in both source and destination clusters...."
	@cat deploy/movepair_controller.yaml | envsubst | kubectl apply -f - --context=$(SRC_CONTEXT)
	@cat deploy/movepair_controller.yaml | envsubst | kubectl apply -f - --context=$(DST_CONTEXT)
	@echo "Done"

# Remove MovePair controller from both source and destination cluster
# Example: make remove_mp_ctrl
.PHONY: remove_mp_ctrl
remove_mp_ctrl:
	@echo "Removing MovePair controller from both source and destination clusters...."
	@kubectl delete -f deploy/movepair_controller.yaml --context=$(SRC_CONTEXT) || true
	@kubectl delete -f deploy/movepair_controller.yaml --context=$(DST_CONTEXT) || true
	@echo "Done"

# Deploy MoveEngine controller in both source and destination cluster
# Example: make deploy_me_ctrl
.PHONY: deploy_me_ctrl
deploy_me_ctrl:
	@echo "Deploying MoveEngine controller in both source and destination clusters...."
	@cat deploy/moveengine_controller.yaml | envsubst | kubectl apply -f - --context=$(SRC_CONTEXT)
	@cat deploy/moveengine_controller.yaml | envsubst | kubectl apply -f - --context=$(DST_CONTEXT)
	@echo "Done"

# Remove MoveEngine controller from both source and destination cluster
# Example: make remove_me_ctrl
.PHONY: remove_me_ctrl
remove_me_ctrl:
	@echo "Removing MoveEngine controller from both source and destination clusters...."
	@kubectl delete -f deploy/moveengine_controller.yaml --context=$(SRC_CONTEXT) || true
	@kubectl delete -f deploy/moveengine_controller.yaml --context=$(DST_CONTEXT) || true
	@echo "Done"

# Deploy DataSync controller in both source and destination cluster
# Example: make deploy_ds_ctrl
.PHONY: deploy_ds_ctrl
deploy_ds_ctrl:
	@echo "Deploying DataSync controller in both source and destination clusters...."
	@cat deploy/datasync_controller.yaml | envsubst | kubectl apply -f - --context=$(SRC_CONTEXT)
	@cat deploy/datasync_controller.yaml | envsubst | kubectl apply -f - --context=$(DST_CONTEXT)
	@echo "Done"

# Remove DataSync controller from both source and destination cluster
# Example: make remove_ds_ctrl
.PHONY: remove_ds_ctrl
remove_ds_ctrl:
	@echo "Removing DataSync controller from both source and destination clusters...."
	@kubectl delete -f deploy/datasync_controller.yaml --context=$(SRC_CONTEXT) || true
	@kubectl delete -f deploy/datasync_controller.yaml --context=$(DST_CONTEXT) || true
	@echo "Done"

# Create a MovePair CR in the source cluster for two local kind clusters
# Example: make create_local_mp
.PHONY: create_local_mp
create_local_mp:
	$(eval SRC_CONTROL_PANE:=$(shell /bin/bash -c "kubectl get pod -n kube-system  -o name --context=$(SRC_CONTEXT)| grep kube-apiserver"))
	$(eval SRC_CLUSTER_IP=$(shell (kubectl get -n kube-system $(SRC_CONTROL_PANE) -o yaml --context=$(SRC_CONTEXT)| grep advertise-address= | cut -c27-)))
	$(eval DST_CONTROL_PANE:=$(shell /bin/bash -c "kubectl get pod -n kube-system  -o name --context=$(DST_CONTEXT)| grep kube-apiserver"))
	$(eval DST_CLUSTER_IP=$(shell (kubectl get -n kube-system $(DST_CONTROL_PANE) -o yaml --context=$(DST_CONTEXT)| grep advertise-address= | cut -c27-)))
	$(eval SRC_CLUSTER_CA:=$(shell (kubectl config view --raw -o=jsonpath='{.clusters[?(@.name=="$(SRC_CONTEXT)")].cluster.certificate-authority-data}')))
	$(eval SRC_CLUSTER_CLIENT_CERT:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(SRC_CONTEXT)")].user.client-certificate-data}')))
	$(eval SRC_CLUSTER_CLIENT_KEY:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(SRC_CONTEXT)")].user.client-key-data}')))
	$(eval DST_CLUSTER_CA:=$(shell (kubectl config view --raw -o=jsonpath='{.clusters[?(@.name=="$(DST_CONTEXT)")].cluster.certificate-authority-data}')))
	$(eval DST_CLUSTER_CLIENT_CERT:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(DST_CONTEXT)")].user.client-certificate-data}')))
	$(eval DST_CLUSTER_CLIENT_KEY:=$(shell (kubectl config view --raw -o=jsonpath='{.users[?(@.name=="$(DST_CONTEXT)")].user.client-key-data}')))

	@echo "Creating local MovePair CR in the source cluster...."
	@/bin/bash -c "                                                 				\
		export SRC_CLUSTER_NAME=$(SRC_CONTEXT) && 									\
		export SRC_CLUSTER_IP=$(SRC_CLUSTER_IP) && 									\
		export SRC_CLUSTER_CA=$(SRC_CLUSTER_CA) && 									\
		export SRC_CLUSTER_CLIENT_CERT=$(SRC_CLUSTER_CLIENT_CERT) &&				\
		export SRC_CLUSTER_CLIENT_KEY=$(SRC_CLUSTER_CLIENT_KEY) &&					\
		export DST_CLUSTER_NAME=$(DST_CONTEXT) && 									\
        export DST_CLUSTER_IP=$(DST_CLUSTER_IP) && 									\
		export DST_CLUSTER_CA=$(DST_CLUSTER_CA) && 									\
		export DST_CLUSTER_CLIENT_CERT=$(DST_CLUSTER_CLIENT_CERT) && 				\
		export DST_CLUSTER_CLIENT_KEY=$(DST_CLUSTER_CLIENT_KEY) &&					\
		cat examples/resources/movepair/local.yaml | envsubst | kubectl apply -f -	\
	"
	@echo "Done"

# Remove local MovePair CR from the source cluster
# Example: make remove_local_mp
.PHONY: remove_local_mp
remove_local_mp:
	@echo "Removing local MovePair CR from controller the source cluster...."
	@kubectl delete -f examples/resources/movepair/local.yaml --context=$(SRC_CONTEXT) || true
	@echo "Done"

# Install necessary Kubemove resources
# Example: make deploy_kubemove
.PHONY: deploy_kubemove
deploy_kubemove: register_crds create_rbac_resources deploy_mp_ctrl deploy_me_ctrl deploy_ds_ctrl

# Remove all kubemove resources from the source and destination clusters
# Example: make purge_kubemove
.PHONY: purge_kubemove
purge_kubemove: remove_ds_ctrl remove_me_ctrl remove_mp_ctrl remove_rbac_resources remove_crds
