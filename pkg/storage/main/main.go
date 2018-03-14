package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	brokerstorage "github.com/nilebox/broker-server/pkg/stateful/storage"
	"github.com/nilebox/brokernetes/pkg/controller/client/typed/brokernetes/v1"
	"github.com/nilebox/brokernetes/pkg/storage"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
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

func main() {
	config, err := ConfigFromEnv()
	if err != nil {
		panic("FUCKING BOLLOCKZ")
	}
	restClient, err := v1.NewForConfig(config)
	if err != nil {
		panic("FUCKING BOLLOCKZ")
	}
	a := storage.NewCrdStorage(restClient, "ivor")
	err = a.CreateInstance(&brokerstorage.InstanceSpec{
		InstanceId: "my-ass",
		Parameters: json.RawMessage("{\"hello world\":\"dog\"}"),
	})
	if err != nil {
		panic(err)
	}

	var input string
	fmt.Scanln(&input)

	err = a.UpdateInstance(&brokerstorage.InstanceSpec{
		InstanceId: "my-ass",
		Parameters: json.RawMessage("{\"hello world\":\"cat\"}"),
	})
	if err != nil {
		panic(err)
	}
}
