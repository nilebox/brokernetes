package storage

import (
	"encoding/json"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/runtime/schema"

	brokerstorage "github.com/nilebox/broker-server/pkg/stateful/storage"
	brokernetesv1 "github.com/nilebox/brokernetes/pkg/apis/brokernetes/v1"
	"github.com/nilebox/brokernetes/pkg/controller/client/typed/brokernetes/v1"
	"github.com/pkg/errors"
)

type crdStorage struct {
	client    v1.OSBInstanceInterface
	namespace string
}

func NewCrdStorage(c *v1.BrokernetesV1Client, namespace string) *crdStorage {
	return &crdStorage{
		client:    c.OSBInstances(namespace),
		namespace: namespace,
	}
}

func ToRawExtension(obj interface{}) (*runtime.RawExtension, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "unexpectedly failed to marshal JSON")
	}
	return &runtime.RawExtension{
		Raw: data,
	}, nil
}

func (s *crdStorage) CreateInstance(instance *brokerstorage.InstanceRecord) error {
	parameters, err := ToRawExtension(instance.Parameters)
	if err != nil {
		return err
	}

	_, err = s.client.Create(&brokernetesv1.OSBInstance{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       brokernetesv1.OSBInstanceResourceKind,
			APIVersion: brokernetesv1.OSBInstanceResourceAPIVersion,
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      instance.InstanceId,
			Namespace: s.namespace,
		},
		Spec: brokernetesv1.OSBInstanceSpec{
			Parameters: parameters,
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *crdStorage) UpdateInstance(instanceId string, parameters json.RawMessage, state brokerstorage.InstanceState) error {
	return nil
}

func (s *crdStorage) UpdateInstanceState(instanceId string, state brokerstorage.InstanceState, err string) error {
	return nil
}

func (s *crdStorage) GetInstance(instanceId string) (*brokerstorage.InstanceRecord, error) {
	return nil, nil
}
