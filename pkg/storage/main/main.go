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
