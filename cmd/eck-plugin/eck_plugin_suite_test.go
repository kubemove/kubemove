package main_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"

	"github.com/kubemove/kubemove/pkg/plugin/proto"
	"google.golang.org/grpc"

	"k8s.io/apimachinery/pkg/util/wait"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	common "github.com/elastic/cloud-on-k8s/pkg/apis/common/v1"
	eck "github.com/elastic/cloud-on-k8s/pkg/apis/elasticsearch/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigPath   string
	srcContext       string
	dstContext       string
	srcPluginAddress string
	dstPluginAddress string

	srcKubeClient kubernetes.Interface
	dstKubeClient kubernetes.Interface
	srcDmClient   dynamic.Interface
	dstDmClient   dynamic.Interface
)

func TestMain(m *testing.M) {
	flag.StringVar(&kubeConfigPath, "kubeconfigpath", "", "KubeConfig path")
	flag.StringVar(&srcContext, "src-context", "", "Source Context")
	flag.StringVar(&dstContext, "dst-context", "", "Destination Context")
	flag.StringVar(&srcPluginAddress, "src-plugin", "", "URL of the source plugin")
	flag.StringVar(&dstPluginAddress, "dst-plugin", "", "URL of the destination plugin")
	flag.Parse()

	os.Exit(m.Run())
}
func TestEckPlugin(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "EckPlugin Suite")
}

var _ = BeforeSuite(func() {
	loader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath}

	srcConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, &clientcmd.ConfigOverrides{CurrentContext: srcContext}).ClientConfig()
	Expect(err).NotTo(HaveOccurred())
	dstConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, &clientcmd.ConfigOverrides{CurrentContext: dstContext}).ClientConfig()
	Expect(err).NotTo(HaveOccurred())

	srcKubeClient, err = kubernetes.NewForConfig(srcConfig)
	Expect(err).NotTo(HaveOccurred())
	srcDmClient, err = dynamic.NewForConfig(srcConfig)
	Expect(err).NotTo(HaveOccurred())

	dstKubeClient, err = kubernetes.NewForConfig(dstConfig)
	Expect(err).NotTo(HaveOccurred())
	dstDmClient, err = dynamic.NewForConfig(dstConfig)
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("ECK Plugin Test", func() {

	Context("Default Nodes", func() {
		It("should sync Elasticsearch data between clusters", func() {

			By("Installing Minio Repository plugin in the source ES")
			err := insertMinioRepository(srcDmClient)
			Expect(err).NotTo(HaveOccurred())

			By("Installing Minio Repository plugin in the destination ES")
			err = insertMinioRepository(dstDmClient)
			Expect(err).NotTo(HaveOccurred())

			By("Establishing gRPC connection with ECK plugin of source cluster")
			srcConn, err := grpc.Dial(srcPluginAddress, grpc.WithInsecure(), grpc.WithBlock())
			Expect(err).NotTo(HaveOccurred())
			defer srcConn.Close()
			srcClient := proto.NewDataSyncerClient(srcConn)

			By("Establishing gRPC connection with ECK plugin of destination cluster")
			dstConn, err := grpc.Dial(dstPluginAddress, grpc.WithInsecure(), grpc.WithBlock())
			Expect(err).NotTo(HaveOccurred())
			defer dstConn.Close()
			dstClient := proto.NewDataSyncerClient(dstConn)

			initParams := &proto.InitRequest{
				Params: map[string]string{
					"engineName":      "sample-es-move",
					"engineNamespace": "default",
				},
			}

			By("Registering Repository in the source Elasticsearch")
			initResp, err := srcClient.Init(context.Background(), initParams)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println(initResp.String())

			By("Registering Repository in the destination Elasticsearch")
			initResp, err = dstClient.Init(context.Background(), initParams)
			Expect(err).NotTo(HaveOccurred())

			snapshotName := fmt.Sprintf("snapshot-%d", time.Now().Unix())
			syncParams := &proto.SyncRequest{
				Params: map[string]string{
					"engineName":      "sample-es-move",
					"engineNamespace": "default",
					"snapshotName":    snapshotName,
				},
			}

			By("Triggering Backup in the source Elasticsearch")
			syncResp, err := srcClient.SyncData(context.Background(), syncParams)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Response: ", syncResp.String())

			statusParams := &proto.SyncStatusRequest{
				Params: map[string]string{
					"engineName":      "sample-es-move",
					"engineNamespace": "default",
					"snapshotName":    snapshotName,
				},
			}

			By("Waiting for Snapshot to be completed")
			err = wait.PollImmediate(2*time.Second, 5*time.Minute, func() (done bool, err error) {
				statusResp, err := srcClient.SyncStatus(context.Background(), statusParams)
				if statusResp == nil {
					return true, err
				}
				fmt.Println("Status: ", statusResp.Status)
				switch statusResp.Status {
				case plugin.Completed:
					return true, nil // snapshot completed successfully. we are done.
				case plugin.InProgress:
					return false, nil // snapshot is running. so, retry.
				default:
					return true, err // snapshot process has encountered a failure. so, no need to retry.
				}
			})
			Expect(err).NotTo(HaveOccurred())

			By("Triggering Restore in the destination Elasticsearch")
			syncResp, err = dstClient.SyncData(context.Background(), syncParams)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Response: ", syncResp.String())

			By("Waiting for Restore to be completed")
			err = wait.PollImmediate(2*time.Second, 5*time.Minute, func() (done bool, err error) {
				statusResp, err := dstClient.SyncStatus(context.Background(), statusParams)
				if statusResp == nil {
					return true, err
				}
				fmt.Println("Status: ", statusResp.Status)
				switch statusResp.Status {
				case plugin.Completed:
					return true, nil // restore completed successfully. we are done.
				case plugin.InProgress:
					return false, nil // restore is running. so, retry.
				default:
					return true, err // restore process has encountered a failure. so, no need to retry.
				}
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func insertMinioRepository(dmClient dynamic.Interface) error {
	gvr := schema.GroupVersionResource{
		Group:    "elasticsearch.k8s.elastic.co",
		Version:  "v1",
		Resource: "elasticsearches",
	}

	es, err := getElasticsearch(dmClient, gvr)
	if err != nil {
		return err
	}

	// insert plugin installer init-container
	es.Spec.NodeSets[0].PodTemplate = corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{
					Name: "install-plugins",
					Command: []string{
						"sh",
						"-c",
						"bin/elasticsearch-plugin install --batch repository-s3",
					},
				},
			},
		},
	}

	// insert minio credentiaals
	es.Spec.SecureSettings = []common.SecretSource{
		{
			SecretName: "minio-credentials",
		},
	}

	// convert updated Elasticsearch back to unstructured object
	updatedES, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&es)
	if err != nil {
		fmt.Println("Error during converting into unstructured object")
		return err
	}

	// update Elasticsearch
	_, err = dmClient.Resource(gvr).Namespace("default").Update(&unstructured.Unstructured{Object: updatedES}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	// give a delay for the ES phase to be updated
	time.Sleep(5 * time.Second)

	// wait for ES to be ready with the plugin installer
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		es, err := getElasticsearch(dmClient, gvr)
		if err == nil {
			return es.Status.Phase == eck.ElasticsearchReadyPhase, nil
		}
		if !kerr.IsNotFound(err) {
			return true, err
		}
		return false, nil
	})
	return err
}

func getElasticsearch(dmClient dynamic.Interface, gvr schema.GroupVersionResource) (*eck.Elasticsearch, error) {
	// read Elasticsearch object
	resp, err := dmClient.Resource(gvr).Namespace("default").Get("sample-es", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// convert to unstructured object into Elasticsearch type
	var es eck.Elasticsearch
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), &es)
	if err != nil {
		return nil, err
	}
	return &es, nil
}
