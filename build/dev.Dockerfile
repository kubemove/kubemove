ARG GO_VERSION
FROM golang:${GO_VERSION}-stretch

ARG GO_PROTO_VERSION
ARG OPERATOR_SDK_VERSION

# install dependencies
RUN set -x \
  && apt-get update \
  && apt-get install -y --no-install-recommends apt-utils ca-certificates wget git bash mercurial bzr xz-utils socat build-essential protobuf-compiler upx \
  && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man /tmp/*

# install protobuffer go-runtime
RUN mkdir -p /go/src/github.com/golang \
  && cd /go/src/github.com/golang \
  && rm -rf protobuf \
  && git clone https://github.com/golang/protobuf.git \
  && cd protobuf \
  && git checkout ${GO_PROTO_VERSION} \
  && GO111MODULE=on go install ./... \
  && cd /go \
  && rm -rf /go/pkg /go/src

# install operator-sdk
RUN set -x \
  && wget -q https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk-${OPERATOR_SDK_VERSION}-x86_64-linux-gnu\
  && mv operator-sdk-${OPERATOR_SDK_VERSION}-x86_64-linux-gnu /usr/bin/operator-sdk \
  && chmod +x /usr/bin/operator-sdk

# install kubectl
RUN set -x                                        \
  && wget -q https://dl.k8s.io/$(curl -fsSL https://storage.googleapis.com/kubernetes-release/release/stable.txt)/kubernetes-client-linux-amd64.tar.gz \
  && tar -xzvf kubernetes-client-linux-amd64.tar.gz \
  && mv kubernetes/client/bin/kubectl /usr/bin/kubectl \
  && chmod +x /usr/bin/kubectl \
  && rm -rf kubernetes kubernetes-client-linux-amd64.tar.gz
