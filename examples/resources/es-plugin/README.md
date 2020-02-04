# Move Elasticsearch Data using Minio Storage

## Prepare Source Cluster

- Get Node Address:

    ```console
    $ kind get kubeconfig --internal --name=src-cluster | grep server
    server: https://172.17.0.2:6443
    ```

- Install ECK Operator
    ```console
    $ kubectl apply -f https://download.elastic.co/downloads/eck/1.0.0/all-in-one.yaml

    customresourcedefinition.apiextensions.k8s.io/apmservers.apm.k8s.elastic.co created
    customresourcedefinition.apiextensions.k8s.io/elasticsearches.elasticsearch.k8s.elastic.co created
    customresourcedefinition.apiextensions.k8s.io/kibanas.kibana.k8s.elastic.co created
    clusterrole.rbac.authorization.k8s.io/elastic-operator created
    clusterrolebinding.rbac.authorization.k8s.io/elastic-operator created
    namespace/elastic-system created
    statefulset.apps/elastic-operator created
    serviceaccount/elastic-operator created
    validatingwebhookconfiguration.admissionregistration.k8s.io/elastic-webhook.k8s.elastic.co created
    service/elastic-webhook-server created
    secret/elastic-webhook-server-cert created
    ```

- Create Minio Credentials Secret
    ```console
    $ kubectl apply -f ./examples/resources/es-plugin/minio-credentials.yaml
    secret/minio-credentials created
    ```

- Create Elasticsearch CR
    ```console
    $ kubectl apply -f ./examples/resources/es-plugin/es.yaml
    elasticsearch.elasticsearch.k8s.elastic.co/sample-es created
    ```

- Get Elasticsearch Credentials
    ```console
    $ kubectl get secret sample-es-es-elastic-user -o=jsonpath='{.data.elastic}' | base64 --decode
    mflmrtm9b8gsp8s95cjztrcl
    ```

- Get HTTP Service
    ```console
    $ kubectl get service sample-es-es-http
    NAME                TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
    sample-es-es-http   NodePort   10.96.201.170   <none>        9200:30045/TCP   36m
    ```

## Prepare Destination Cluster

- Get Node Address:

    ```console
    $ kind get kubeconfig --internal --name=dst-cluster | grep server
    server: https://172.17.0.3:6443
    ```

- Install ECK Operator
    ```console
    kubectl apply -f https://download.elastic.co/downloads/eck/1.0.0/all-in-one.yam
    ```

- Create Minio Credentials Secret
    ```console
    kubectl apply -f ./examples/resources/es-plugin/minio-credentials.yaml
    ```

- Create Elasticsearch CR
    ```console
    kubectl apply -f ./examples/resources/es-plugin/es.yaml
    ```
- Get Elasticsearch Credentials
    ```console
    $ kubectl get secret sample-es-es-elastic-user -o=jsonpath='{.data.elastic}' | base64 --decode
    bm44hfqncclslvw88dvq266v
    ```

- Get Elasticsearch Service NodePort
    ```console
    $ kubectl get service sample-es-es-http
    NAME                TYPE       CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
    sample-es-es-http   NodePort   10.96.59.64   <none>        9200:31222/TCP   36m
    ```

- Deploy Minio Server
    ```console
    $ kubectl apply -f ./examples/resources/es-plugin/minio.yaml
    persistentvolumeclaim/minio-pvc created
    deployment.apps/minio-server created
    service/minio created
    ```

    ```console
    kubectl get service minio
    NAME    TYPE       CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
    minio   NodePort   10.96.5.241   <none>        9000:32304/TCP   4m2s
    ```

## Backup

- Register Snapshot Repository

    ```console
    curl -u "elastic:mflmrtm9b8gsp8s95cjztrcl" -k -X PUT "https://172.17.0.2:30045/_snapshot/minio_repository?pretty" -H 'Content-Type: application/json' -d'
    {
    "type": "s3",
    "settings": {
        "bucket": "demo",
        "base_path": "default/sample-es",
        "endpoint": "172.17.0.3:32304",
        "protocol": "http"
    }
    }
    '
    ```

    ```json
    {
    "acknowledged" : true
    }
    ```

- Insert a Index

    ```console
    curl -u "elastic:mflmrtm9b8gsp8s95cjztrcl" -k -X PUT "https://172.17.0.2:30045/twitter?pretty"
    ```

    ```json
    {
    "acknowledged" : true,
    "shards_acknowledged" : true,
    "index" : "twitter"
    }
    ```

- Take Snapshot

    ```console
    curl -u "elastic:mflmrtm9b8gsp8s95cjztrcl" -k -X PUT "https://172.17.0.2:30045/_snapshot/minio_repository/snapshot_1?pretty"
    ```

    ```json
    {
    "accepted" : true
    }
    ```

- Verify Snapshot has Succeeded

    ```console
    curl -u "elastic:mflmrtm9b8gsp8s95cjztrcl" -k -X GET "https://172.17.0.2:30045/_snapshot/minio_repository/snapshot_1?pretty"
    ```

    ```json
    {
    "snapshots" : [
        {
        "snapshot" : "snapshot_1",
        "uuid" : "wC7ncjxdSuyyOyypTbRLOw",
        "version_id" : 7050299,
        "version" : "7.5.2",
        "indices" : [
            "twitter"
        ],
        "include_global_state" : true,
        "state" : "SUCCESS",
        "start_time" : "2020-02-04T12:42:18.335Z",
        "start_time_in_millis" : 1580820138335,
        "end_time" : "2020-02-04T12:42:18.536Z",
        "end_time_in_millis" : 1580820138536,
        "duration_in_millis" : 201,
        "failures" : [ ],
        "shards" : {
            "total" : 1,
            "failed" : 0,
            "successful" : 1
        }
        }
    ]
    }
    ```

## Restore

- Register Snapshot Repository

    ```console
    curl -u "elastic:bm44hfqncclslvw88dvq266v" -k -X PUT "https://172.17.0.3:31222/_snapshot/minio_repository?pretty" -H 'Content-Type: application/json' -d'
    {
    "type": "s3",
    "settings": {
        "bucket": "demo",
        "base_path": "default/sample-es",
        "endpoint": "172.17.0.3:32304",
        "protocol": "http",
        "read_only": "true"
    }
    }
    '
    ```

    ```json
    {
    "acknowledged" : true
    }
    ```

- Close all indexes

    ```console
    curl -u "elastic:bm44hfqncclslvw88dvq266v" -k -X POST "https://172.17.0.3:31222/_all/_close?pretty"
    ```

    ```json
    {
    "acknowledged" : true,
    "shards_acknowledged" : false,
    "indices" : { }
    }
    ```

- Restore from Snapshot

    ```console
    curl -u "elastic:bm44hfqncclslvw88dvq266v" -k -X POST "https://172.17.0.3:31222/_snapshot/minio_repository/snapshot_1/_restore?pretty"
    ```

    ```json
    {
    "accepted" : true
    }
    ```

- Check restore succeeded

    ```console
    curl -u "elastic:bm44hfqncclslvw88dvq266v" -k -X GET "https://172.17.0.3:31222/_recovery?pretty"
    ```

    ```json
    {
    "twitter" : {
        "shards" : [
        {
            "id" : 0,
            "type" : "SNAPSHOT",
            "stage" : "DONE",
            "primary" : true,
            "start_time_in_millis" : 1580820518633,
            "stop_time_in_millis" : 1580820518982,
            "total_time_in_millis" : 349,
            "source" : {
            "repository" : "minio_repository",
            "snapshot" : "snapshot_1",
            "version" : "7.5.2",
            "index" : "twitter",
            "restoreUUID" : "eJr4PS_VTt60ay3p6F33ug"
            },
            "target" : {
            "id" : "8zZiTcMgScONNDvNI3aWDQ",
            "host" : "10.244.0.24",
            "transport_address" : "10.244.0.24:9300",
            "ip" : "10.244.0.24",
            "name" : "sample-es-es-default-1"
            },
            "index" : {
            "size" : {
                "total_in_bytes" : 230,
                "reused_in_bytes" : 0,
                "recovered_in_bytes" : 230,
                "percent" : "100.0%"
            },
            "files" : {
                "total" : 1,
                "reused" : 0,
                "recovered" : 1,
                "percent" : "100.0%"
            },
            "total_time_in_millis" : 159,
            "source_throttle_time_in_millis" : 0,
            "target_throttle_time_in_millis" : 0
            },
            "translog" : {
            "recovered" : 0,
            "total" : 0,
            "percent" : "100.0%",
            "total_on_start" : 0,
            "total_time_in_millis" : 101
            },
            "verify_index" : {
            "check_index_time_in_millis" : 0,
            "total_time_in_millis" : 0
            }
        },
        {
            "id" : 0,
            "type" : "PEER",
            "stage" : "DONE",
            "primary" : false,
            "start_time_in_millis" : 1580820519180,
            "stop_time_in_millis" : 1580820519625,
            "total_time_in_millis" : 444,
            "source" : {
            "id" : "8zZiTcMgScONNDvNI3aWDQ",
            "host" : "10.244.0.24",
            "transport_address" : "10.244.0.24:9300",
            "ip" : "10.244.0.24",
            "name" : "sample-es-es-default-1"
            },
            "target" : {
            "id" : "yhAN0T-cQjaRb47NQkSjNw",
            "host" : "10.244.0.23",
            "transport_address" : "10.244.0.23:9300",
            "ip" : "10.244.0.23",
            "name" : "sample-es-es-default-0"
            },
            "index" : {
            "size" : {
                "total_in_bytes" : 230,
                "reused_in_bytes" : 0,
                "recovered_in_bytes" : 230,
                "percent" : "100.0%"
            },
            "files" : {
                "total" : 1,
                "reused" : 0,
                "recovered" : 1,
                "percent" : "100.0%"
            },
            "total_time_in_millis" : 167,
            "source_throttle_time_in_millis" : 0,
            "target_throttle_time_in_millis" : 0
            },
            "translog" : {
            "recovered" : 0,
            "total" : 0,
            "percent" : "100.0%",
            "total_on_start" : 0,
            "total_time_in_millis" : 156
            },
            "verify_index" : {
            "check_index_time_in_millis" : 0,
            "total_time_in_millis" : 0
            }
        }
        ]
    }
    }
    ```

    or

    ```console
    curl -u "elastic:bm44hfqncclslvw88dvq266v" -k -X GET "https://172.17.0.3:31222/_cat/recovery?v&h=i,st,snap"

    ```

    ```console
    i       st   snap
    twitter done snapshot_1
    twitter done n/a
    ```

- Verify restored data

    ```console
    curl -u "elastic:bm44hfqncclslvw88dvq266v" -k -X GET "https://172.17.0.3:31222/twitter?pretty"
    ```

    ```json
    {
    "twitter" : {
        "aliases" : { },
        "mappings" : { },
        "settings" : {
        "index" : {
            "creation_date" : "1580820012761",
            "number_of_shards" : "1",
            "number_of_replicas" : "1",
            "uuid" : "XYCeS2FqQACjZDOZU1lPLA",
            "version" : {
            "created" : "7050299"
            },
            "provided_name" : "twitter"
        }
        }
    }
    }
    ```
