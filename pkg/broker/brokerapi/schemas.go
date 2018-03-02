package brokerapi

import "encoding/json"

// Schemas represents a broker's schemas for both service instances and service
// bindings
type Schemas struct {
	Instance ServiceInstanceSchema `json:"service_instance"`
	Binding  ServiceBindingSchema  `json:"service_binding"`
}

type ServiceInstanceSchema struct {
	Create Schema `json:"create"`
	Update Schema `json:"update"`
}

type ServiceBindingSchema struct {
	Create Schema `json:"create"`
}

// Schema consists of the schema for inputs and the schema for outputs.
// Schemas are in the form of JSON Schema v4 (http://json-schema.org/).
type Schema struct {
	Parameters json.RawMessage `json:"parameters"`
}
