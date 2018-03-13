package v1

import (
	"bytes"
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	InstanceResourceSingular = "osbinstance"
	InstanceResourcePlural   = "osbinstances"
	InstanceResourceVersion  = "v1"
	InstanceResourceKind     = "OsbInstance"

	// TODO should be dynamic (specified by the user), at least the prefix
	// GroupName is the group name use in this package.
	GroupName = "brokernetes.nilebox.github.com"

	// TODO GroupName should be dynamic
	InstanceResourceAPIVersion = GroupName + "/" + InstanceResourceVersion
	InstanceResourceName       = InstanceResourcePlural + "." + GroupName
)

type InstanceConditionType string

// These are valid conditions of a Instance object.
const (
	InstanceInProgress InstanceConditionType = "InProgress"
	InstanceReady      InstanceConditionType = "Ready"
	InstanceError      InstanceConditionType = "Error"
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

// +genclient
// +genclient:noStatus

// Instance is handled by Instance controller.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Instance struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Instance.
	Spec InstanceSpec `json:"spec,omitempty"`

	// Most recently observed status of the Instance.
	Status InstanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=true
type InstanceSpec struct {
	// TODO add necessary fields there

	// TODO encrypt parameters or use Secret
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`
	// TODO encrypt outputs or use Secret
	Output *runtime.RawExtension `json:"output,omitempty"`
}

// +k8s:deepcopy-gen=true
// InstanceCondition describes the state of a Instance object at a certain point.
type InstanceCondition struct {
	// Type of Instance condition.
	Type InstanceConditionType `json:"type"`
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

func (sc *InstanceCondition) String() string {
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
type InstanceStatus struct {
	// Represents the latest available observations of a Instance's current state.
	Conditions []InstanceCondition `json:"conditions,omitempty"`
}

func (ss *InstanceStatus) String() string {
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

func (s *Instance) GetCondition(conditionType InstanceConditionType) (int, *InstanceCondition) {
	for i, condition := range s.Status.Conditions {
		if condition.Type == conditionType {
			return i, &condition
		}
	}
	return -1, nil
}

// Updates existing Instance condition or creates a new one. Sets LastTransitionTime to now if the
// status has changed.
// Returns true if Instance condition has changed or has been added.
func (s *Instance) UpdateCondition(condition *InstanceCondition) bool {
	cond := *condition // copy to avoid mutating the original
	now := meta_v1.Now()
	cond.LastTransitionTime = now
	// Try to find this Instance condition.
	conditionIndex, oldCondition := s.GetCondition(cond.Type)

	if oldCondition == nil {
		// We are adding new Instance condition.
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

// InstanceList is a list of Instances.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type InstanceList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata,omitempty"`

	Items []Instance `json:"items"`
}
