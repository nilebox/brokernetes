package storage

import (
	"encoding/json"

	brokerstorage "github.com/nilebox/broker-server/pkg/stateful/storage"
	"k8s.io/client-go/rest"
)

type crdStorage struct {
}

func (s *crdStorage) CreateInstance(instance *brokerstorage.InstanceRecord) error {
	rest
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
