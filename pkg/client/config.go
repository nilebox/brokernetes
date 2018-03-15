package client

import (
	"net"
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func ConfigFromEnv() (*rest.Config, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return nil, errors.New("unable to load cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}
	CAFile, CertFile, KeyFile := os.Getenv("KUBERNETES_CA_PATH"), os.Getenv("KUBERNETES_CLIENT_CERT"), os.Getenv("KUBERNETES_CLIENT_KEY")
	if CAFile == "" || CertFile == "" || KeyFile == "" {
		return nil, errors.New("unable to load TLS configuration, KUBERNETES_CA_PATH, KUBERNETES_CLIENT_CERT and KUBERNETES_CLIENT_KEY must be defined")
	}
	return &rest.Config{
		Host: "https://" + net.JoinHostPort(host, port),
		TLSClientConfig: rest.TLSClientConfig{
			CAFile:   CAFile,
			CertFile: CertFile,
			KeyFile:  KeyFile,
		},
	}, nil
}

func ConfigFromFile(configFileName, configContext string) (*rest.Config, error) {
	var configApi *clientcmdapi.Config
	configApi, err := clientcmd.LoadFromFile(configFileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load REST client configuration from file %q", configFileName)
	}
	return clientcmd.NewDefaultClientConfig(*configApi, &clientcmd.ConfigOverrides{
		CurrentContext: configContext,
	}).ClientConfig()
}

func InClusterConfig() (*rest.Config, error) {
	return rest.InClusterConfig()
}
