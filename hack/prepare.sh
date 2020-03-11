#!/usr/bin/env bash
set -eou pipefail

echo "Registering crds in the source cluster"
kubectl apply -f ./deploy/crds --context=${SRC_CLUSTER_NAME}
echo "Registering crds in the destination cluster"
kubectl apply -f ./deploy/crds --context=${DST_CLUSTER_NAME}

echo ""
echo "Creating RBAC resources in the source cluster"
kubectl apply -f ./deploy/rbac.yaml --context=${SRC_CLUSTER_NAME}
echo "Creating RBAC resources in the destination cluster"
kubectl apply -f ./deploy/rbac.yaml --context=${DST_CLUSTER_NAME}

echo""
echo "Deploying MovePair controller in the source cluster"
kubectl apply -f ./deploy/movepair_controller.yaml --context=${SRC_CLUSTER_NAME}
kubectl wait --for=condition=READY pods --all  --namespace=kubemove --timeout=5m --context=${SRC_CLUSTER_NAME} || true

echo ""
echo "Creating MovePair in the source cluster"
cat ./examples/resources/movepair/local.yaml | envsubst | kubectl apply -f - --context=${SRC_CLUSTER_NAME}
