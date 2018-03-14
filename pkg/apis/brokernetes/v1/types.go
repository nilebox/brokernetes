package v1

import (
	"bytes"
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/nilebox/brokernetes/pkg/apis/brokernetes"
)

const (
	OSBInstanceResourceSingular = "osbinstance"
	OSBInstanceResourcePlural   = "osbinstances"
	OSBInstanceResourceVersion  = "v1"
	OSBInstanceResourceKind     = "OSBInstance"

	// TODO GroupName should be dynamic
	OSBInstanceResourceAPIVersion = brokernetes.GroupName + "/" + OSBInstanceResourceVersion
	OSBInstanceResourceName       = OSBInstanceResourcePlural + "." + brokernetes.GroupName
)

type OSBInstanceConditionType string

// These are valid conditions of a OSBInstance object.
const (
	OSBInstanceInProgress OSBInstanceConditionType = "InProgress"
	OSBInstanceReady      OSBInstanceConditionType = "Ready"
	OSBInstanceError      OSBInstanceConditionType = "Error"
)

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type OSBInstanceOperationType string

const (
	OperationCreate OSBInstanceOperationType = "Create"
	OperationUpdate OSBInstanceOperationType = "Update"
	OperationDelete OSBInstanceOperationType = "Delete"
)

// +genclient
// +genclient:noStatus

// OSBInstance is handled by OSBInstance controller.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type OSBInstance struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the OSBInstance.
	Spec OSBInstanceSpec `json:"spec,omitempty"`

	// Most recently observed status of the OSBInstance.
	Status OSBInstanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=true
type OSBInstanceSpec struct {
	// TODO add necessary fields there
	ServiceId string `json:"serviceId"`
	PlanId    string `json:"planId"`

	// TODO encrypt parameters or use Secret
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`
	// TODO encrypt outputs or use Secret
	Output *runtime.RawExtension `json:"output,omitempty"`
}

// +k8s:deepcopy-gen=true
// OSBInstanceCondition describes the state of a OSBInstance object at a certain point.
type OSBInstanceCondition struct {
	// Type of OSBInstance condition.
	Type OSBInstanceConditionType `json:"type"`
	// Status of the condition.
	Status ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime meta_v1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime meta_v1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

func (sc *OSBInstanceCondition) String() string {
	var buf bytes.Buffer
	buf.WriteString(string(sc.Type))
	buf.WriteByte(' ')
	buf.WriteString(string(sc.Status))
	if sc.Reason != "" {
		fmt.Fprintf(&buf, " %q", sc.Reason)
	}
	if sc.Message != "" {
		fmt.Fprintf(&buf, " %q", sc.Message)
	}
	return buf.String()
}

// +k8s:deepcopy-gen=true
type OSBInstanceStatus struct {
	// Represents the latest available observations of a OSBInstance's current state.
	Conditions        []OSBInstanceCondition   `json:"conditions,omitempty"`
	LastOperationType OSBInstanceOperationType `json:"lastOperationType,omitempty"`
	Error             string                   `json:"error"`
}

func (ss *OSBInstanceStatus) String() string {
	first := true
	var buf bytes.Buffer
	buf.WriteByte('[')
	for _, cond := range ss.Conditions {
		if first {
			first = false
		} else {
			buf.WriteByte('|')
		}
		buf.WriteString(cond.String())
	}
	buf.WriteByte(']')
	return buf.String()
}

func (s *OSBInstance) GetCondition(conditionType OSBInstanceConditionType) (int, *OSBInstanceCondition) {
	for i, condition := range s.Status.Conditions {
		if condition.Type == conditionType {
			return i, &condition
		}
	}
	return -1, nil
}

// Updates existing OSBInstance condition or creates a new one. Sets LastTransitionTime to now if the
// status has changed.
// Returns true if OSBInstance condition has changed or has been added.
func (s *OSBInstance) UpdateCondition(condition *OSBInstanceCondition) bool {
	cond := *condition // copy to avoid mutating the original
	now := meta_v1.Now()
	cond.LastTransitionTime = now
	// Try to find this OSBInstance condition.
	conditionIndex, oldCondition := s.GetCondition(cond.Type)

	if oldCondition == nil {
		// We are adding new OSBInstance condition.
		s.Status.Conditions = append(s.Status.Conditions, cond)
		return true
	}
	// We are updating an existing condition, so we need to check if it has changed.
	if cond.Status == oldCondition.Status {
		cond.LastTransitionTime = oldCondition.LastTransitionTime
	}

	isEqual := cond.Status == oldCondition.Status &&
		cond.Reason == oldCondition.Reason &&
		cond.Message == oldCondition.Message &&
		cond.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

	if !isEqual {
		cond.LastUpdateTime = now
	}

	s.Status.Conditions[conditionIndex] = cond
	// Return true if one of the fields have changed.
	return !isEqual
}

func (s *OSBInstance) GetLastOperationType() OSBInstanceOperationType {
	return s.Status.LastOperationType
}

func (s *OSBInstance) UpdateLastOperationType(lastOperationType OSBInstanceOperationType) {
	s.Status.LastOperationType = lastOperationType
}

// OSBInstanceList is a list of OSBInstances.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type OSBInstanceList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata,omitempty"`

	Items []OSBInstance `json:"items"`
}
