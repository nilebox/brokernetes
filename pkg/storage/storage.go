package storage

import (
	"bytes"
	"encoding/json"
	"errors"

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	brokerstorage "github.com/nilebox/broker-server/pkg/stateful/storage"
	brokernetesv1 "github.com/nilebox/brokernetes/pkg/apis/brokernetes/v1"
	"github.com/nilebox/brokernetes/pkg/controller/client/typed/brokernetes/v1"
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

// CreateInstance for crdStorage just stores the instance parameters
func (s *crdStorage) CreateInstance(instance *brokerstorage.InstanceSpec) error {
	parameters := &runtime.RawExtension{
		Raw: []byte(instance.Parameters),
	}

	inProgressCond := &brokernetesv1.OSBInstanceCondition{Type: brokernetesv1.OSBInstanceInProgress, Status: brokernetesv1.ConditionTrue}
	readyCond := &brokernetesv1.OSBInstanceCondition{Type: brokernetesv1.OSBInstanceReady, Status: brokernetesv1.ConditionFalse}
	errorCond := &brokernetesv1.OSBInstanceCondition{Type: brokernetesv1.OSBInstanceError, Status: brokernetesv1.ConditionFalse}

	osbInstance := &brokernetesv1.OSBInstance{
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
	}
	osbInstance.UpdateCondition(inProgressCond)
	osbInstance.UpdateCondition(readyCond)
	osbInstance.UpdateCondition(errorCond)
	osbInstance.UpdateLastOperationType(brokernetesv1.OperationCreate)

	_, err := s.client.Create(osbInstance)
	if err != nil {
		// TODO handle the conflict case
		return err
	}
	return nil
}

func compareParameters(old *brokernetesv1.OSBInstance, instanceParameters *brokerstorage.InstanceSpec) bool {
	// TODO stupid assumption: because we merely copy the bytes into RawExtension, this ... is probably the same thing?
	return bytes.Equal(old.Spec.Parameters.Raw, []byte(instanceParameters.Parameters))
}

func (s *crdStorage) update(instanceId string, f func(*brokernetesv1.OSBInstance)) error {
	for {
		instance, err := s.client.Get(instanceId, meta_v1.GetOptions{})
		if err != nil {
			// TODO handle instance not found
			return err
		}

		if instance.DeletionTimestamp != nil {
			return brokerstorage.NewDeleting("instance is already being deleted")
		}

		// Check for in_progress
		index, condition := instance.GetCondition(brokernetesv1.OSBInstanceInProgress)
		if index == -1 {
			return errors.New("unexpectedly missing condition")
		}
		if condition.Status == brokernetesv1.ConditionTrue {
			return brokerstorage.NewInProgress("checking for in_progress before updating CRD failed")
		}

		// Update kubernetes
		osbInstance := brokernetesv1.OSBInstance{}
		instance.DeepCopyInto(&osbInstance)

		f(&osbInstance)

		_, err = s.client.Update(&osbInstance)
		if err != nil {
			if api_errors.IsConflict(err) {
				// Retry
				continue
			}
			return err
		}

		return nil
	}
}

func (s *crdStorage) UpdateInstance(updated *brokerstorage.InstanceSpec) error {
	return s.update(updated.InstanceId, func(existing *brokernetesv1.OSBInstance) {
		// Check for idempotency first - if idempotent, do nothing
		if compareParameters(existing, updated) {
			return
		}

		// Replace the existing parameters
		existing.Spec.Parameters = &runtime.RawExtension{
			Raw: []byte(updated.Parameters),
		}

		// Set the status to UPDATE_IN_PROGRESS
		inProgressCond := &brokernetesv1.OSBInstanceCondition{Type: brokernetesv1.OSBInstanceInProgress, Status: brokernetesv1.ConditionTrue}
		readyCond := &brokernetesv1.OSBInstanceCondition{Type: brokernetesv1.OSBInstanceReady, Status: brokernetesv1.ConditionFalse}
		errorCond := &brokernetesv1.OSBInstanceCondition{Type: brokernetesv1.OSBInstanceError, Status: brokernetesv1.ConditionFalse}

		existing.UpdateCondition(inProgressCond)
		existing.UpdateCondition(readyCond)
		existing.UpdateCondition(errorCond)
		existing.UpdateLastOperationType(brokernetesv1.OperationUpdate)
	})
}

func (s *crdStorage) DeleteInstance(instanceId string) error {
	return s.update(instanceId, func(existing *brokernetesv1.OSBInstance) {
		// Set the status to DELETE_IN_PROGRESS
		inProgressCond := &brokernetesv1.OSBInstanceCondition{Type: brokernetesv1.OSBInstanceInProgress, Status: brokernetesv1.ConditionTrue}
		readyCond := &brokernetesv1.OSBInstanceCondition{Type: brokernetesv1.OSBInstanceReady, Status: brokernetesv1.ConditionFalse}
		errorCond := &brokernetesv1.OSBInstanceCondition{Type: brokernetesv1.OSBInstanceError, Status: brokernetesv1.ConditionFalse}

		existing.UpdateCondition(inProgressCond)
		existing.UpdateCondition(readyCond)
		existing.UpdateCondition(errorCond)
		existing.UpdateLastOperationType(brokernetesv1.OperationDelete)
	})
}

func (s *crdStorage) GetInstance(instanceId string) (*brokerstorage.InstanceRecord, error) {
	instance, err := s.client.Get(instanceId, meta_v1.GetOptions{})
	if err != nil {
		// TODO handle instance not found
		return nil, err
	}

	osbState, err := getState(instance)
	if err != nil {
		return nil, err
	}

	return &brokerstorage.InstanceRecord{
		Spec: brokerstorage.InstanceSpec{
			InstanceId: instance.GetName(),
			ServiceId:  instance.Spec.ServiceId,
			PlanId:     instance.Spec.PlanId,
			Parameters: json.RawMessage(instance.Spec.Parameters.Raw),
			Outputs:    json.RawMessage(instance.Spec.Output.Raw),
		},
		State: osbState,
		Error: instance.Status.Error,
	}, nil
}

func getState(instance *brokernetesv1.OSBInstance) (brokerstorage.InstanceState, error) {
	eint, errorCondition := instance.GetCondition(brokernetesv1.OSBInstanceError)
	if eint == -1 {
		return "", errors.New("error condition is not found")
	}
	pint, inProgressCondition := instance.GetCondition(brokernetesv1.OSBInstanceInProgress)
	if pint == -1 {
		return "", errors.New("in progress condition is not found")
	}
	rint, readyCondition := instance.GetCondition(brokernetesv1.OSBInstanceReady)
	if rint == -1 {
		return "", errors.New("ready condition is not found")
	}

	var suffix string
	if errorCondition.Status == brokernetesv1.ConditionTrue {
		suffix = "Failed"
	} else if readyCondition.Status == brokernetesv1.ConditionTrue {
		suffix = "Succeeded"
	} else if inProgressCondition.Status == brokernetesv1.ConditionTrue {
		suffix = "InProgress"
	} else {
		return "", errors.New("invalid condition state")
	}

	var result string
	switch instance.GetLastOperationType() {
	case brokernetesv1.OperationCreate:
		result = "Create" + suffix
	case brokernetesv1.OperationDelete:
		result = "Delete" + suffix
	case brokernetesv1.OperationUpdate:
		result = "Update" + suffix
	}

	return brokerstorage.InstanceState(result), nil
}
