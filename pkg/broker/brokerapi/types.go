package brokerapi

// Types represents the types offered by a given service plan-- instances
// and/or bindings
type Types struct {
	Instance string `json:"instance"`
	Binding  string `json:"binding"`
}

const (
	// InstanceType is a string constant representation of the instance type
	InstanceType = "instanceType"
	// BindingType is a string constant representation of the binding type
	BindingType = "bindingType"
)
