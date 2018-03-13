// Generated file, do not modify manually!

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/nilebox/brokernetes/pkg/controller/client/typed/brokernetes/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeBrokernetesV1 struct {
	*testing.Fake
}

func (c *FakeBrokernetesV1) OSBInstances(namespace string) v1.OSBInstanceInterface {
	return &FakeOSBInstances{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeBrokernetesV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
