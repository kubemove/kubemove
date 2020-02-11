package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
