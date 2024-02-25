package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	runclient "github.com/fluxcd/pkg/runtime/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TLSClientConfig struct {
	Insecure   bool   `json:"insecure"`
	CertData   string `json:"certData"`
	KeyData    string `json:"keyData"`
	ServerName string `json:"serverName"`
}

type SecretConfig struct {
	TLSClientConfig TLSClientConfig `json:"tlsClientConfig"`
}

type ClusterConfig struct {
	ExternalAddress string
	InternalAddress string
	Name            string
	Server          string
	TLSClientConfig TLSClientConfig
}

func KubeClientForLeafCluster(mgmt client.Client, clusterName string, opts *runclient.Options) (client.Client, *ClusterConfig, error) {
	secretName := fmt.Sprintf("%s-cluster", clusterName)
	secret := corev1.Secret{}
	err := mgmt.Get(context.Background(), client.ObjectKey{Namespace: "argocd", Name: secretName}, &secret)
	if err != nil {
		return nil, nil, err
	}

	// Parse the config block from the secret
	configData, ok := secret.Data["config"]
	if !ok {
		return nil, nil, fmt.Errorf("config block not found in secret")
	}

	var secretConfig SecretConfig
	err = json.Unmarshal(configData, &secretConfig)
	if err != nil {
		return nil, nil, err
	}

	clusterConfig := &ClusterConfig{}
	clusterConfig.ExternalAddress = secret.Annotations["flamingo/external-address"]
	clusterConfig.InternalAddress = secret.Annotations["flamingo/internal-address"]
	clusterConfig.Name = string(secret.Data["name"])
	clusterConfig.Server = string(secret.Data["server"])
	clusterConfig.TLSClientConfig = secretConfig.TLSClientConfig

	certData, err := base64.StdEncoding.DecodeString(clusterConfig.TLSClientConfig.CertData)
	if err != nil {
		return nil, nil, err
	}
	keyData, err := base64.StdEncoding.DecodeString(clusterConfig.TLSClientConfig.KeyData)
	if err != nil {
		return nil, nil, err
	}

	// Manually construct the Kubernetes rest.Config
	cfg := &rest.Config{
		// Set the server address
		Host: clusterConfig.ExternalAddress,
		// Setup custom TLS config
		TLSClientConfig: rest.TLSClientConfig{
			Insecure:   clusterConfig.TLSClientConfig.Insecure,
			CertData:   certData,
			KeyData:    keyData,
			ServerName: clusterConfig.TLSClientConfig.ServerName,
		},
		QPS:   opts.QPS,
		Burst: opts.Burst,
	}

	// Ensure your scheme is properly configured with the required API types
	k8sClient, err := client.New(cfg, client.Options{Scheme: NewScheme()})
	if err != nil {
		return nil, clusterConfig, err
	}

	return k8sClient, clusterConfig, nil
}
