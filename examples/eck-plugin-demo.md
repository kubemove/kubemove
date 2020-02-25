## Prepare Clusters

- Set VM heap max size

```console
sudo sysctl -w vm.max_map_count=262144
```

- Register CRDs in source cluster

```console
kubectl apply -f ./deploy/crds/ --context=kind-src-cluster=
```

- Register CRDs in destination cluster

```console
kubectl apply -f ./deploy/crds/ --context=kind-dst-cluster=
```

- Deploy ECK Operator in source cluster

```console
kubectl apply -f https://download.elastic.co/downloads/eck/1.0.1/all-in-one.yaml --context=kind-src-cluster
```

- Deploy ECK Operator in destination cluster

```console
kubectl apply -f https://download.elastic.co/downloads/eck/1.0.1/all-in-one.yaml --context=kind-dst-cluster
```

## Prepare Backend

- Create Minio credentials in source cluster
```console
kubectl apply -f ./examples/resources/es-plugin/minio-credentials.yaml --context=kind-src-cluster
```

- Create Minio credentials in destination cluster
```console
kubectl apply -f ./examples/resources/es-plugin/minio-credentials.yaml --context=kind-dst-cluster
```

- Deploy Minio Server in destination cluster
```console
kubectl apply -f ./examples/resources/es-plugin/minio.yaml --context=kind-dst-cluster
```
- Create Bucket `demo` in the Minio server
```console
kubectl get services minio --context=kind-dst-cluster
NAME    TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
minio   NodePort   10.96.27.155   <none>        9000:32559/TCP   54m
       9200/TCP         44m
```


## Prepare Elasticsearch

- Deploy ES in source cluster
```console
kubectl apply -f ./examples/resources/es-plugin/es.yaml --context=kind-src-cluster
```

- Deploy ES in destination cluster
```console
kubectl apply -f ./examples/resources/es-plugin/es.yaml --context=kind-src-cluster
```

- Get ES credentials
```console
$ kubectl get secret sample-es-es-elastic-user -o=jsonpath='{.data.elastic}' --context=kind-src-cluster | base64 --decode
  qx46p9gxcl5lqrs5cl9g6b94
```
- Portforward ES http service
```console
kubectl port-forward service/sample-es-es-http 9200 --context=kind-src-cluster
```

- Create sample data in the source ES

```console
curl -u "elastic:qx46p9gxcl5lqrs5cl9g6b94" -k -X PUT "https://localhost:9200/kubemove?pretty"
```


## Deploy Plugin

- Deploy plugin in the source cluster

```console
kubectl apply -f ./examples/resources/es-plugin/eck-plugin.yaml --context=kind-src-cluster
```

- Deploy plugin in the destination cluster

```console
kubectl apply -f ./examples/resources/es-plugin/eck-plugin.yaml --context=kind-dst-cluster
```

## Prepare for Data Sync

- Deploy MoveEngine in the source cluster
```console
kubectl apply -f ./examples/resources/es-plugin/engine.yaml --context=kind-src-cluster
```

- Deploy MoveEngine in the destination cluster
```console
kubectl apply -f ./examples/resources/es-plugin/engine.yaml --context=kind-dst-cluster
```