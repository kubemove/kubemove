#!/usr/bin/env bash
set -eou pipefail

echo "Registering CRDs in the source cluster"
kubectl apply -f deploy/crds --context=${SRC_CLUSTER_NAME}
echo "Registering CRDs in the destination cluster"
kubectl apply -f deploy/crds --context=${DST_CLUSTER_NAME}

echo ""
echo "Creating RBAC resources in the source cluster"
kubectl apply -f deploy/rbac.yaml --context=${SRC_CLUSTER_NAME}
echo "Creating RBAC resources in the destination cluster"
kubectl apply -f deploy/rbac.yaml --context=${DST_CLUSTER_NAME}

echo""
echo "Deploying MovePair controller in the source cluster"
cat deploy/movepair_controller.yaml | envsubst | kubectl apply -f - --context=${SRC_CLUSTER_NAME}

echo ""
echo "Deploying DataSync controller in the source cluster"
cat deploy/datasync_controller.yaml | envsubst | kubectl apply -f - --context=${SRC_CLUSTER_NAME}
echo "Deploying DataSync controller in the destination cluster"
cat deploy/datasync_controller.yaml | envsubst | kubectl apply -f - --context=${DST_CLUSTER_NAME}

echo ""
echo "Deploying MoveEngine controller in the source cluster"
cat deploy/moveengine_controller.yaml | envsubst | kubectl apply -f - --context=${SRC_CLUSTER_NAME}
echo "Deploying MoveEngine controller in the destination cluster"
cat deploy/moveengine_controller.yaml | envsubst | kubectl apply -f - --context=${DST_CLUSTER_NAME}

echo ""
echo "Creating Minio Secret in the source cluster"
cat deploy/minio_secret.yaml | envsubst | kubectl apply -f - --context=${SRC_CLUSTER_NAME}
echo "Creating Minio Secret in the destination cluster"
cat deploy/minio_secret.yaml | envsubst | kubectl apply -f - --context=${DST_CLUSTER_NAME}

echo ""
echo "Deploying Minio Server in the destination cluster"
kubectl apply -f deploy/minio_server.yaml --context=${DST_CLUSTER_NAME}

echo ""
echo "Waiting for all pods of of the source cluster to be ready"
kubectl wait --for=condition=READY pods --all --timeout=5m --context=${SRC_CLUSTER_NAME} || true
echo "Waiting for all pods of of the source cluster to be ready"
kubectl wait --for=condition=READY pods --all --timeout=5m --context=${DST_CLUSTER_NAME} || true

echo ""
echo "Creating MovePair in the source cluster"
cat examples/resources/movepair/local.yaml | envsubst | kubectl apply -f - --context=${SRC_CLUSTER_NAME}

echo ""
echo "Creating demo bucket in the  Minio server"
MINIO_NODEPORT=$(kubectl get service minio -o yaml --context=${DST_CLUSTER_NAME} | grep nodePort | cut -c15-)
MINIO_SERVER_ADDRESS=${DST_CLUSTER_IP}:${MINIO_NODEPORT}

sleep 60
mc config host add es-repo http://${MINIO_SERVER_ADDRESS} ${MINIO_ACCESS_KEY} ${MINIO_SECRET_KEY}
mc mb es-repo/demo
