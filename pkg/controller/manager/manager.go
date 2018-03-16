package manager

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/nilebox/broker-server/pkg/stateful/task"
	osb_v1 "github.com/nilebox/brokernetes/pkg/apis/brokernetes/v1"
	"github.com/nilebox/brokernetes/pkg/controller/client/typed/brokernetes/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	DefaultNumWorkers = 20

	FinalizerName = "osb_worker.brokernetes/finalizer"
)

type Manager struct {
	client v1.OSBInstanceInterface
	broker task.Broker

	// ASSSUMPTION: brokers are idempotent - if we want to reduce duplicate calls to broker we might want this...
	// This tracks the current in-progress instance IDs, so on resyncs we don't
	// make boatloads of new EC2 instances
	// processing map[string]bool

	// This is the work queue
	// TODO: implement work queue
}


func NewManager(client v1.OSBInstanceInterface, broker task.Broker) *Manager {
	return &Manager{
		broker: broker,
		client: client,
	}
}

// Run spins up the manager worker pool
func (m *Manager) Run() {
}

// Process adds an instance to the pool for processing
func (m *Manager) Process(instance *osb_v1.OSBInstance) (bool, error) {
	// Before any of this happens, add a finalizer if instance doesn't have one
	found := false
	for _, f := range instance.GetFinalizers() {
		if f == FinalizerName {
			found = true
		}
	}
	if !found {
		instance.Finalizers = append(instance.Finalizers, FinalizerName)
		m.client.Update(instance)
		return false, nil
	}

	// If the instance is IN_PROGRESS we should try to deal with it
	i, inProgress := instance.GetCondition(osb_v1.OSBInstanceInProgress)
	if i == -1 {
		return false, errors.New("Condition not found lol")
	}

	if inProgress.Status != osb_v1.ConditionTrue {
		// Nothing to do
		return false, nil
	}

	// Switch between create update and delete
	switch instance.GetLastOperationType() {
	case osb_v1.OperationCreate:
		return m.processCreate(instance)
	case osb_v1.OperationUpdate:
		return m.processUpdate(instance)
	case osb_v1.OperationDelete:
		return m.processDelete(instance)
	default:
		return false, errors.New("SHIT IS FUUUUUUCKKKKKKKEDDDDD'[:w")
	}
}

func (m *Manager) processCreate(instance *osb_v1.OSBInstance) (bool, error) {
	// TODO check the last modified (lease) in case of multiple controllers...
	message, err := m.broker.CreateInstance(instance.GetName(), json.RawMessage(instance.Spec.Parameters.Raw))
	err = m.handleConditions(instance, err)
	if err != nil {
		return false, err
	}
	instance.Spec.Output = &runtime.RawExtension{
		Raw: []byte(message),
	}
	instance.UpdateLastOperationType(osb_v1.OperationCreate)
	_, err = m.client.Update(instance)
	if err != nil {
		return true, err
	}
	return false, nil
}

func (m *Manager) handleConditions(instance *osb_v1.OSBInstance, err error) (error) {
	var conditionChanged bool
	if err != nil {
		errorCond := &osb_v1.OSBInstanceCondition{Type: osb_v1.OSBInstanceError, Status: osb_v1.ConditionTrue}
		conditionChanged = instance.UpdateCondition(errorCond)
		instance.Status.Error = err.Error()
	} else {
		readyCond := &osb_v1.OSBInstanceCondition{Type: osb_v1.OSBInstanceReady, Status: osb_v1.ConditionTrue}
		conditionChanged = instance.UpdateCondition(readyCond)
	}

	inProgressCond := &osb_v1.OSBInstanceCondition{Type: osb_v1.OSBInstanceInProgress, Status: osb_v1.ConditionFalse}
	instance.UpdateCondition(inProgressCond)

	if !conditionChanged {
		return errors.New("unexpected lack of condition change.")
	}
	return nil
}

func (m *Manager) processDelete(instance *osb_v1.OSBInstance) (bool, error) {
	// TODO check the last modified (lease) in case of multiple controllers...
	err := m.broker.DeleteInstance(instance.GetName(), json.RawMessage(instance.Spec.Parameters.Raw))
	err = m.handleConditions(instance, err)
	if err != nil {
		return false, err
	}
	instance.UpdateLastOperationType(osb_v1.OperationDelete)
	_, err = m.client.Update(instance)
	if err != nil {
		return true, err
	}
	return false, nil
}

func (m *Manager) processUpdate(instance *osb_v1.OSBInstance) (bool, error) {
	// TODO check the last modified (lease) in case of multiple controllers...
	message, err := m.broker.UpdateInstance(instance.GetName(), json.RawMessage(instance.Spec.Parameters.Raw))
	err = m.handleConditions(instance, err)
	if err != nil {
		return false, err
	}
	instance.Spec.Output = &runtime.RawExtension{
		Raw: []byte(message),
	}
	instance.UpdateLastOperationType(osb_v1.OperationCreate)
	_, err = m.client.Update(instance)
	if err != nil {
		return true, err
	}
	return false, nil
}
