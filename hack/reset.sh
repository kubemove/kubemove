#!/usr/bin/env bash
set -eou pipefail

echo ""
echo "Removing all DataSync CRs from the source cluster"
kubectl delete datasync --all-namespaces --all --context=${SRC_CLUSTER_NAME} || true
echo "Removing all DataSync CRs from the destination cluster"
kubectl delete datasync --all-namespaces --all  --context=${DST_CLUSTER_NAME} || true

echo ""
echo "Removing all MoveEngine CRs from the source cluster"
kubectl delete moveengine --all-namespaces --all --context=${SRC_CLUSTER_NAME} || true
echo "Removing all MoveEngine CRs from the destination cluster"
kubectl delete moveengine --all-namespaces --all  --context=${DST_CLUSTER_NAME} || true

echo ""
echo "Removing MovePair from the source cluster"
kubectl delete -f  examples/resources/movepair/local.yaml --context=${SRC_CLUSTER_NAME} || true

echo""
echo "Removing MovePair controller from the source cluster"
kubectl delete -f  deploy/movepair_controller.yaml  --context=${SRC_CLUSTER_NAME} || true

echo ""
echo "Removing DataSync controller from the source cluster"
kubectl delete -f deploy/datasync_controller.yaml --context=${SRC_CLUSTER_NAME} || true
echo "Removing DataSync controller from the destination cluster"
kubectl delete -f deploy/datasync_controller.yaml --context=${DST_CLUSTER_NAME} || true

echo ""
echo "Removing MoveEngine controller from the source cluster"
kubectl delete -f  deploy/moveengine_controller.yaml --context=${SRC_CLUSTER_NAME} || true
echo "Removing MoveEngine controller from the destination cluster"
kubectl delete -f  deploy/moveengine_controller.yaml --context=${DST_CLUSTER_NAME} || true

echo ""
echo "Removing Minio Secret from the source cluster"
kubectl delete -f  deploy/minio_secret.yaml --context=${SRC_CLUSTER_NAME} || true
echo "Removing Minio Secret from the destination cluster"
kubectl delete -f  deploy/minio_secret.yaml --context=${DST_CLUSTER_NAME} || true

echo ""
echo "Removing Minio Server from the destination cluster"
kubectl delete -f deploy/minio_server.yaml --context=${DST_CLUSTER_NAME} || true

echo ""
echo "Removing RBAC resources from the source cluster"
kubectl delete -f deploy/rbac.yaml --context=${SRC_CLUSTER_NAME} || true
echo "Removing RBAC resources from the destination cluster"
kubectl delete -f deploy/rbac.yaml --context=${DST_CLUSTER_NAME} || true

echo ""
echo "Removing CRDs from the source cluster"
kubectl delete -f deploy/crds --context=${SRC_CLUSTER_NAME} || true
echo "Removing CRDs from the destination cluster"
kubectl delete -f deploy/crds --context=${DST_CLUSTER_NAME} || true

echo ""
echo "Waiting for all pods of of the source cluster to be ready"
kubectl wait --for=condition=READY pods --all --timeout=5m --context=${SRC_CLUSTER_NAME} || true
echo "Waiting for all pods of of the source cluster to be ready"
kubectl wait --for=condition=READY pods --all --timeout=5m --context=${DST_CLUSTER_NAME} || true
