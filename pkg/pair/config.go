package mpair

import (
	"context"

	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdapilatest "k8s.io/client-go/tools/clientcmd/api/latest"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func FetchPairClient(mpair *v1alpha1.MovePair) (client.Client, error) {
	config, err := loadClientCmdConfig(mpair.Spec.Config)
	if err != nil {
		return nil, err
	}

	client, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func FetchPairDiscoveryClient(mpair *v1alpha1.MovePair) (*discovery.DiscoveryClient, error) {
	config, err := loadClientCmdConfig(mpair.Spec.Config)
	if err != nil {
		return nil, err
	}

	return discovery.NewDiscoveryClientForConfigOrDie(config), nil
}

func FetchDiscoveryClient() (*discovery.DiscoveryClient, error) {
	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	return discovery.NewDiscoveryClientForConfigOrDie(config), nil
}

func FetchPairDynamicClient(mpair *v1alpha1.MovePair) (dynamic.Interface, error) {
	config, err := loadClientCmdConfig(mpair.Spec.Config)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

// TODO export it
func loadClientCmdConfig(v1Config clientcmdapiv1.Config) (*rest.Config, error) {
	var config clientcmdapi.Config

	err := clientcmdapilatest.Scheme.Convert(&v1Config, &config, context.TODO())
	if err != nil {
		return nil, err
	}

	return clientcmd.NewNonInteractiveClientConfig(
		config,
		config.CurrentContext,
		&clientcmd.ConfigOverrides{},
		clientcmd.NewDefaultClientConfigLoadingRules()).
		ClientConfig()
}

func Validate(mpair *v1alpha1.MovePair) error {
	var config clientcmdapi.Config

	err := clientcmdapilatest.Scheme.Convert(&mpair.Spec.Config, &config, context.TODO())
	if err != nil {
		return err
	}

	return clientcmd.Validate(config)
}
