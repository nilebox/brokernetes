package controller

import (
	"context"

	"github.com/nilebox/brokernetes/pkg/broker/brokerapi"
)

// Controller defines the APIs that all controllers are expected to support. Implementations
// should be concurrency-safe
type Controller interface {
	Catalog(ctx context.Context) (*brokerapi.Catalog, error)

	GetServiceInstanceStatus(ctx context.Context, instanceID, serviceID, planID, operation string) (*brokerapi.GetServiceInstanceStatusResponse, error)
	CreateServiceInstance(ctx context.Context, instanceID string, acceptsIncomplete bool, req *brokerapi.CreateServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error)
	UpdateServiceInstance(ctx context.Context, instanceID string, acceptsIncomplete bool, req *brokerapi.UpdateServiceInstanceRequest) (*brokerapi.UpdateServiceInstanceResponse, error)
	RemoveServiceInstance(ctx context.Context, instanceID, serviceID, planID string, acceptsIncomplete bool) (*brokerapi.DeleteServiceInstanceResponse, error)

	Bind(ctx context.Context, instanceID, bindingID string, req *brokerapi.BindingRequest) (*brokerapi.CreateServiceBindingResponse, error)
	UnBind(ctx context.Context, instanceID, bindingID, serviceID, planID string) error
}
