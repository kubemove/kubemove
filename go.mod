module github.com/kubemove/kubemove

go 1.13

require (
	cloud.google.com/go v0.39.0 // indirect
	github.com/coreos/prometheus-operator v0.29.0 // indirect
	github.com/elastic/cloud-on-k8s v0.0.0-20200225093456-57e96b20f060
	github.com/elastic/go-elasticsearch/v7 v7.5.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/spec v0.19.2
	github.com/golang/protobuf v1.3.2
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/operator-framework/operator-sdk v0.7.0
	github.com/pkg/errors v0.8.1
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5
	github.com/vmware-tanzu/velero v1.2.0
	google.golang.org/grpc v1.27.1
	k8s.io/api v0.0.0-20191114100352-16d7abae0d2a
	k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
	k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0

)
