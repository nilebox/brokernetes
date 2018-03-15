package manager

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/nilebox/broker-server/pkg/stateful/task"
	osb_v1 "github.com/nilebox/brokernetes/pkg/apis/brokernetes/v1"
	"github.com/nilebox/brokernetes/pkg/controller/client/typed/brokernetes/v1"
)

const (
	DefaultNumWorkers = 20

	FinalizerName = "osb_worker"
)

type Manager struct {
	client v1.OSBInstanceInterface
	broker task.Broker

	// This tracks the current in-progress instance IDs, so on resyncs we don't
	// make boatloads of new EC2 instances
	processing map[string]bool

	// This is the work queue
	// TODO: implement work queue
}

// Run spins up the manager worker pool
func (m *Manager) Run() {
}

// Process adds an instance to the pool for processing
func (m *Manager) Process(instance *osb_v1.OSBInstance) (bool, error) {
	// Before any of this happens, add a finalizer if instance doesn't have one
	found := false
	for f := range instance.GetFinalizers() {
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
	case osb_v1.OperationUpdate:
	case osb_v1.OperationDelete:
	default:
		return false, errors.New("SHIT IS FUUUUUUCKKKKKKKEDDDDD'[:w")
	}
}

func (m *manager) processOperationHelper(instance *osb_v1.OSBInstance, f func(*osb_v1.OSBInstance) error) (bool, error) {
	if m.processing[instance.GetName()] {
		// This is already in progress, supposedly. Do nothing.
		// TODO: death of a goroutine
		return false, nil
	}

	err := f(instance)
	if err != nil {
	}

	// 	response, err := m.broker.CreateInstance(instance.GetName(), json.RawMessage(instance.Spec.Parameters.Raw))
	// 	if err != nil {
	// 		// The create has failed
	// 		// TODO: deal with other cases, conflict, etc,
	// 		// Need to add a custom type

	// 		// Unset the processing state
	// 	}

	// 	// Create success, store the instance back into kubernetes
	// 	// Hopefully nothing else has touched it...

	// Unset the processing state
}

func (m *Manager) processCreate(instance *osb_v1.OSBInstance) (bool, error) {
}

func (m *Manager) processDelete(instance *osb_v1.OSBInstance) (bool, error) {
	if m.processing[instance.GetName()] {
		// This is already in progress, supposedly. Do nothing.
		// TODO: death of a goroutine
		return false, nil
	}

	err := m.broker.DeleteInstance(instance.GetName(), json.RawMessage(instance.Spec.Parameters.Raw))
	if err != nil {
		// The delete has failed
		// TODO: deal with other cases, conflict, etc that might arise from delete?
		// Need to add a custom type

		// Unset the processing state
	}

	// Delete success, delete the instance from kubernetes (? concurrency?)
	// Hopefully nothing else has touched it...

	// Unset the processing state (probably delete it from the map)

}

func (m *Manager) processUpdate(instance *osb_v1.OSBInstance) (bool, error) {
	if m.processing[instance.GetName()] {
		// This is already in progress, supposedly. Do nothing.
		// TODO: death of a goroutine
		return false, nil
	}

	err := m.broker.UpdateInstance(instance.GetName(), json.RawMessage(instance.Spec.Parameters.Raw))
	if err != nil {
		// The delete has failed
		// TODO: deal with other cases, conflict, etc that might arise from delete?
		// Need to add a custom type

		// Unset the processing state
	}

	// Delete success, delete the instance from kubernetes (? concurrency?)
	// Hopefully nothing else has touched it...

	// Unset the processing state (probably delete it from the map)

}
