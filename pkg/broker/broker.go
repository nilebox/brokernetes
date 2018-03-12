package broker

import (
	"encoding/json"
	brokerapi "github.com/nilebox/broker-server/pkg/api"
	"go.uber.org/zap"

	"github.com/nilebox/broker-server/pkg/stateful/task"
)

const (
	// TODO nilebox: support multiple services and plans
	ServiceId   = "uuid1"
	ServiceName = "my-service"
	PlanId      = "uuid2"
	PlanName    = "default"
)

type crdBroker struct {
	log *zap.Logger
}

func NewController(log *zap.Logger) (task.Broker, error) {
	return &crdBroker{
		log: log,
	}, nil
}

func (c *crdBroker) Catalog() (*brokerapi.Catalog, error) {
	c.log.Info("Catalog called")

	catalog := brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				ID:          ServiceId,
				Name:        ServiceName,
				Description: "Service description",
				Bindable:    true,
				Plans: []brokerapi.ServicePlan{
					{
						ID:          PlanId,
						Name:        PlanName,
						Description: "Plan description",
						//Schemas: &brokerapi.Schemas{
						//	Instance: brokerapi.InstanceSchema{
						//		Create: brokerapi.Schema{
						//			Parameters: c.schema,
						//		},
						//		Update: brokerapi.Schema{
						//			Parameters: c.schema,
						//		},
						//	},
						//},
					},
				},
				PlanUpdateable: false,
			},
		},
	}

	return &catalog, nil
}

func (c *crdBroker) CreateInstance(instanceId string, parameters json.RawMessage) (json.RawMessage, error) {
	return nil, nil
}

func (c *crdBroker) UpdateInstance(instanceId string, parameters json.RawMessage) (json.RawMessage, error) {
	return nil, nil
}

func (c *crdBroker) DeleteInstance(instanceId string, parameters json.RawMessage) error {
	return nil
}
