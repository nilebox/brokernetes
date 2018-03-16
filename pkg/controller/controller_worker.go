package controller

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"

	osb_v1 "github.com/nilebox/brokernetes/pkg/apis/brokernetes/v1"
)

const (
	// maxRetries is the number of times a OSBInstance object will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
)

func (c *Controller) worker(ctx context.Context) {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	retriable, err := c.processKey(key.(string))
	c.handleErr(retriable, err, key)

	return true
}

func (c *Controller) handleErr(retriable bool, err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}
	if retriable && c.queue.NumRequeues(key) < maxRetries {
		log.Printf("Error syncing OSBInstance %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	log.Printf("Dropping OSBInstance %q out of the queue: %v", key, err)
	c.queue.Forget(key)
}

func (c *Controller) processKey(key string) (bool /*retriable*/, error) {
	startTime := time.Now()
	log.Printf("Started syncing OSBInstance %q", key)
	defer func() {
		log.Printf("Finished syncing OSBInstance %q (%v)", key, time.Now().Sub(startTime))
	}()
	osbInstanceObj, exists, err := c.osbInstanceInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return false, errors.WithStack(err)
	}
	if !exists {
		log.Printf("OSBInstance %q has been deleted", key)
		return false, nil
	}

	// Deep-copy otherwise we are mutating our cache.
	instance := osbInstanceObj.(*osb_v1.OSBInstance).DeepCopy()

	// TODO: might not need these fields
	instance.APIVersion = osb_v1.OSBInstanceResourceAPIVersion
	instance.Kind = osb_v1.OSBInstanceResourceKind

	// Do something with our state
	//log.Printf("Doing sometihng")

	// TODO: this should probably end up kicking off a gofunc. We'll mark the
	//       queue item as "done". Eventually this gofunc should end up
	//       updating the state item as it's last thing that it does
	c.manager.Process(instance)

	// conflict, retriable, bundle, err := c.process(instance)
	// if conflict {
	// 	return false, nil
	// }
	// conflict, retriable, err = c.handleProcessResult(instance, bundle, retriable, err)
	// if conflict {
	// 	return false, nil
	// }
	// return retriable, err
	return false, nil
}

// func (c *Controller) process(state *orch_v1.State) (conflictRet, retriableRet bool, b *smith_v1.Bundle, e error) {
// 	// Grab the namespace
// 	namespace, exists, err := c.namespaceInformer.GetIndexer().GetByKey(state.Namespace)
// 	if err != nil {
// 		return false, false, nil, errors.WithStack(err)
// 	}
// 	if !exists {
// 		return false, false, nil, errors.Errorf("missing namespace %q in informer", state.Namespace)
// 	}

// 	// Grab the ConfigMap
// 	key := byConfigMapNameIndexKey(state.Namespace, state.Spec.ConfigMapName)
// 	if key == "" {
// 		return false, false, nil, errors.Errorf("configMapName is not provided in state spec for %q", state.GetName())
// 	}
// 	configMap, exists, err := c.configMapInformer.GetIndexer().GetByKey(key)
// 	if err != nil {
// 		return false, false, nil, errors.WithStack(err)
// 	}
// 	if !exists {
// 		return false, false, nil, errors.Errorf("missing ConfigMap %q (key: %q) in informer", state.Spec.ConfigMapName, key)
// 	}

// 	// Entangle the state, passing in the namespace and and configmap as context
// 	entanglerContext := &EntanglerContext{
// 		Config: configMap.(*core_v1.ConfigMap).Data,
// 		Label:  GetNamespaceLabel(namespace.(*core_v1.Namespace)),
// 	}
// 	bundleSpec, retriable, err := c.entangler.Entangle(state, entanglerContext)
// 	if err != nil {
// 		return false, retriable, nil, fmt.Errorf("failed to wire up Bundle for State %q: %v", state.Name, err)
// 	}
// 	conflict, retriable, bundle, err := c.createOrUpdateBundle(state, bundleSpec)
// 	if err != nil {
// 		return false, retriable, nil, fmt.Errorf("failed to create/update Bundle for State %q: %v", state.Name, err)
// 	}
// 	if conflict {
// 		return true, false, nil, nil
// 	}
// 	return false, false, bundle, nil
// }

// func (c *Controller) createOrUpdateBundle(state *orch_v1.State, bundleSpec *smith_v1.Bundle) (conflictRet, retriableRet bool, b *smith_v1.Bundle, e error) {
// 	existingObj, exists, err := c.bundleInformer.GetIndexer().Get(bundleSpec)
// 	if err != nil {
// 		return false, false, nil, err
// 	}
// 	if exists {
// 		bundle := existingObj.(*smith_v1.Bundle).DeepCopy()
// 		bundle.SetGroupVersionKind(smith_v1.BundleGVK) // Typed objects have their TypeMeta erased. Put it back.
// 		return c.updateBundle(state, bundleSpec, bundle)
// 	}
// 	return c.createBundle(bundleSpec)
// }

// func (c *Controller) createBundle(bundleSpec *smith_v1.Bundle) (conflictRet, retriableRet bool, b *smith_v1.Bundle, e error) {
// 	result, err := c.bundleClient.Bundles(bundleSpec.Namespace).Create(bundleSpec)
// 	if err != nil {
// 		if api_errors.IsAlreadyExists(err) {
// 			return true, false, nil, nil // State will be requeued because of the concurrent create
// 		}
// 		return false, true, nil, err
// 	}
// 	return false, false, result, nil
// }

// func (c *Controller) updateBundle(state *orch_v1.State, bundleSpec, existingBundle *smith_v1.Bundle) (conflictRet, retriableRet bool, b *smith_v1.Bundle, e error) {
// 	if existingBundle.DeletionTimestamp != nil {
// 		return false, false, nil, fmt.Errorf("bundle %q is marked for deletion", existingBundle.Name)
// 	}
// 	if !meta_v1.IsControlledBy(existingBundle, state) {
// 		return false, false, nil, fmt.Errorf("bundle %q is not owned by the State", existingBundle.Name)
// 	}

// 	updated, match, err := c.specCheck.CompareActualVsSpec(bundleSpec, existingBundle)
// 	if err != nil {
// 		return false, false, nil, fmt.Errorf("error comparing spec and actual Bundle: %v", err)
// 	}
// 	if match {
// 		log.Printf("Bundle %s/%s exists and is up to date", bundleSpec.Namespace, bundleSpec.Name)
// 		return false, false, existingBundle, nil
// 	}
// 	var updatedBundle smith_v1.Bundle
// 	err = runtime.DefaultUnstructuredConverter.FromUnstructured(updated.Object, &updatedBundle)
// 	if err != nil {
// 		return false, false, nil, err
// 	}
// 	log.Printf("Bundle %s/%s exists but does not match the spec, updating", bundleSpec.Namespace, bundleSpec.Name)
// 	result, err := c.bundleClient.Bundles(bundleSpec.Namespace).Update(&updatedBundle)
// 	if err != nil {
// 		if api_errors.IsConflict(err) {
// 			return true, false, nil, nil // State will be requeued because of the concurrent update
// 		}
// 		return false, true, nil, err
// 	}
// 	return false, false, result, nil
// }

// func (c *Controller) handleProcessResult(state *orch_v1.State, bundle *smith_v1.Bundle, retriable bool, err error) (conflictRet, retriableRet bool, e error) {
// 	if err == context.Canceled || err == context.DeadlineExceeded {
// 		return false, false, nil
// 	}

// 	inProgressCond := orch_v1.StateCondition{
// 		Type:   orch_v1.StateInProgress,
// 		Status: orch_v1.ConditionFalse,
// 	}
// 	readyCond := orch_v1.StateCondition{
// 		Type:   orch_v1.StateReady,
// 		Status: orch_v1.ConditionFalse,
// 	}
// 	errorCond := orch_v1.StateCondition{
// 		Type:   orch_v1.StateError,
// 		Status: orch_v1.ConditionFalse,
// 	}

// 	if err != nil {
// 		errorCond.Status = orch_v1.ConditionTrue
// 		errorCond.Message = err.Error()
// 		if retriable {
// 			errorCond.Reason = "RetriableError"
// 			inProgressCond.Status = orch_v1.ConditionTrue
// 		} else {
// 			errorCond.Reason = "TerminalError"
// 		}
// 	} else if len(bundle.Status.Conditions) == 0 {
// 		// smith is not currently reporting any Conditions;
// 		// presumably we've just created something.
// 		inProgressCond.Status = orch_v1.ConditionTrue
// 		inProgressCond.Reason = "WaitingOnSmithConditions"
// 		inProgressCond.Message = "Waiting for Smith to report Conditions (initial creation?)"
// 	} else {
// 		// map bundle statuses that we understand.
// 		smithToVoyagerStateMapping := map[smith_v1.BundleConditionType]*orch_v1.StateCondition{
// 			smith_v1.BundleInProgress: &inProgressCond,
// 			smith_v1.BundleReady:      &readyCond,
// 			smith_v1.BundleError:      &errorCond,
// 		}

// 		for bundleCondType, currentCond := range smithToVoyagerStateMapping {
// 			_, bundleCond := bundle.GetCondition(bundleCondType)

// 			if bundleCond == nil {
// 				currentCond.Status = orch_v1.ConditionUnknown
// 				currentCond.Reason = "SmithInteropError"
// 				currentCond.Message = "Smith not reporting state for this condition"
// 				continue
// 			}

// 			if bundleCond.Reason != "" {
// 				currentCond.Reason = "Smith" + bundleCond.Reason
// 			}
// 			if bundleCond.Message != "" {
// 				currentCond.Message = "Smith: " + bundleCond.Message
// 			}
// 			switch bundleCond.Status {
// 			case smith_v1.ConditionTrue:
// 				currentCond.Status = orch_v1.ConditionTrue
// 			case smith_v1.ConditionUnknown:
// 				currentCond.Status = orch_v1.ConditionUnknown
// 			}
// 		}
// 	}

// 	inProgressUpdated := state.UpdateCondition(&inProgressCond)
// 	readyUpdated := state.UpdateCondition(&readyCond)
// 	errorUpdated := state.UpdateCondition(&errorCond)

// 	// Updating the State status
// 	if inProgressUpdated || readyUpdated || errorUpdated {
// 		conflictStatus, retriableStatus, errStatus := c.setStatus(state)
// 		if errStatus != nil {
// 			if err != nil {
// 				log.Printf("%v", errStatus)
// 				return false, retriableStatus || retriable, err
// 			}
// 			return false, retriableStatus, errStatus
// 		}
// 		if conflictStatus {
// 			return true, false, nil
// 		}
// 	}
// 	return false, retriable, err
// }

// func (c *Controller) setStatus(state *orch_v1.State) (conflictRet, retriableRet bool, e error) {
// 	log.Printf("Setting State status to %s", &state.Status)
// 	state, err := c.stateClient.States(state.Namespace).Update(state)
// 	if err != nil {
// 		if api_errors.IsConflict(err) {
// 			return true, false, nil
// 		}
// 		return false, true, fmt.Errorf("failed to set State status: %v", err)
// 	}
// 	return false, false, nil
// }
