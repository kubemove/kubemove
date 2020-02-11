package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"

	es "github.com/elastic/go-elasticsearch/v7"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *eckDDM) newElasticsearchClient(opt ElasticsearchOptions) (*es.Client, error) {
	// configure client
	cfg := es.Config{
		Addresses: []string{fmt.Sprintf("%s://%s", opt.Scheme, opt.ServiceName)},
	}

	// configure authentication
	err := d.configureBasicAuth(&cfg, opt)
	if err != nil {
		return nil, err
	}

	if opt.Scheme == "https" {
		err = d.configureTLS(&cfg, opt)
	}

	return es.NewClient(cfg)
}

func (d *eckDDM) configureBasicAuth(cfg *es.Config, opt ElasticsearchOptions) error {
	// get the elastic user secret
	authSecret, err := d.k8sClient.CoreV1().Secrets(opt.Namespace).Get(opt.AuthSecret, metav1.GetOptions{})
	if err != nil {
		return err
	}

	password, ok := authSecret.Data[ElasticUser]
	if !ok {
		return fmt.Errorf("failed to set Authorization Headers. Reason: no passowrd found for %q user", ElasticUser)
	}

	cfg.Username = ElasticUser
	cfg.Password = string(password)

	return nil
}

func (d *eckDDM) configureTLS(cfg *es.Config, opt ElasticsearchOptions) error {
	// get the internal cert secret
	certSecret, err := d.k8sClient.CoreV1().Secrets(opt.Namespace).Get(opt.TLSSecret, metav1.GetOptions{})
	if err != nil {
		return err
	}

	caCert, found := certSecret.Data[TLSCertKey]
	if !found {
		return fmt.Errorf("failed to set TLS transport. Reason: tls.crt not found in secret %s/%s", opt.Namespace, opt.TLSSecret)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
	cfg.Transport = tr

	return nil
}

func (d *eckDDM) extractPluginParameters(params map[string]interface{}) (PluginParameters, string, error) {
	parameters := PluginParameters{}

	val, found := params[MoveEngineName]
	if !found {
		return parameters, "", fmt.Errorf("failed to extract plugin's parameters. Reason: MoveEngine name not found in the paramters")
	}
	engineName := val.(string)

	val, found = params[MoveEngineNamespace]
	if !found {
		return parameters, "", fmt.Errorf("failed to extract plugin's parameters. Reason: MoveEngine namespace not found in the paramters")
	}
	engineNamespace := val.(string)

	// read MoveEngine CR as unstructured object using dynamic client
	moveEngineGVR := v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralMoveEngine)
	resp, err := d.dmClient.Resource(moveEngineGVR).Namespace(engineNamespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return parameters, "", err
	}

	// convert unstructured object into MoveEngine CR
	var moveEngine v1alpha1.MoveEngine
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), &moveEngine)
	if err != nil {
		return parameters, "", err
	}

	// extract repository parameters
	repoParams, found := moveEngine.Spec.PluginParameters[RepositoryParameters]
	if !found {
		return parameters, "", fmt.Errorf("repository parameters not found in pluginParameters")
	}
	parameters.Repository, found = repoParams.(RepositoryOptions)
	if !found {
		return parameters, "", fmt.Errorf("failed to cast repository prameters into RepositoryOptions")
	}

	// extract elasticsearch parameters
	esParams, found := moveEngine.Spec.PluginParameters[ElasticsearchParameter]
	if !found {
		return parameters, "", fmt.Errorf("elasticsearch parameters not found in pluginParameters")
	}
	parameters.Elasticsearch, found = esParams.(ElasticsearchOptions)
	if !found {
		return parameters, "", fmt.Errorf("failed to cast elasticsearch prameters into ElasticsearchOptions")
	}

	return parameters, moveEngine.Spec.Mode, nil
}
