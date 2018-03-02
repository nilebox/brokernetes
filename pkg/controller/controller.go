package controller

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/nilebox/brokernetes/pkg/broker/brokerapi"
	"github.com/nilebox/brokernetes/pkg/broker/controller"
	"github.com/nilebox/brokernetes/pkg/util/zappers"
	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

const (
	// TODO nilebox: support multiple services and plans
	ServiceId   = "uuid1"
	ServiceName = "my-service"
	PlanId      = "uuid2"
	PlanName    = "default"

	StatusInProgress = "in progress"
	StatusSucceeded  = "succeeded"
	StatusFailed     = "failed"

	// The operation is sent from the API as a string "{operation}:{arn}"
	OperationCreate = "create"
	OperationDelete = "delete"
	OperationUpdate = "update"
)

type brokernetesController struct {
	appCtx    context.Context
	schema    json.RawMessage
	validator *gojsonschema.Schema
}

func NewController(appCtx context.Context) (*brokernetesController, error) {
	//schema, err := schemas.InstanceSchema()
	//if err != nil {
	//	return nil, err
	//}
	//loader := gojsonschema.NewGoLoader(schema)
	//validator, err := gojsonschema.NewSchema(loader)
	//if err != nil {
	//	return nil, errors.Wrap(err, "failure to create schema from loader")
	//}

	return &brokernetesController{
		appCtx: appCtx,
		//schema:    schema,
		//validator: validator,
	}, nil
}

func validatePlan(serviceId string, planId string) error {
	if serviceId != ServiceId {
		return errors.Errorf("unexpected ServiceId %s expected %s", serviceId, ServiceId)
	}

	if planId != PlanId {
		return errors.Errorf("unexpected PlanId %s expected %s", planId, PlanId)
	}

	return nil
}

func handleServiceError(err error) error {
	// TODO
	return err
}

func (c *brokernetesController) Catalog(ctx context.Context) (*brokerapi.Catalog, error) {
	log := c.appCtx.Value("log").(*zap.Logger)
	log.Info("Catalog called")

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
						Schemas: &brokerapi.Schemas{
							Instance: brokerapi.ServiceInstanceSchema{
								Create: brokerapi.Schema{
									Parameters: c.schema,
								},
								Update: brokerapi.Schema{
									Parameters: c.schema,
								},
							},
						},
					},
				},
				PlanUpdateable: false,
			},
		},
	}

	return &catalog, nil
}

func validateSchema(validator *gojsonschema.Schema, value json.RawMessage) error {
	loader := gojsonschema.NewGoLoader(value)
	validationResult, err := validator.Validate(loader)
	if err != nil {
		return errors.Wrap(err, "parameters did not validate against schema")
	}
	if !validationResult.Valid() {
		validationErrors := validationResult.Errors()
		msgs := make([]string, 0, len(validationErrors))

		for _, validationErr := range validationErrors {
			msgs = append(msgs, validationErr.String())
		}

		return errors.New("could not validate json: " + strings.Join(msgs, ", "))
	}

	return nil
}

func (c *brokernetesController) CreateServiceInstance(ctx context.Context, instanceID string, acceptsIncomplete bool, req *brokerapi.CreateServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error) {

	log := ctx.Value("log").(*zap.Logger)
	log = log.With(zappers.InstanceID(instanceID))
	log.Info("CreateServiceInstance called")

	// There should only be one of this running at any one time, and any prior failures should be cleaned up by
	// service catalog, but who knows. Designed this to be slightly more resilient against dumb usage.
	err := validatePlan(req.ServiceID, req.PlanID)
	if err != nil {
		return nil, controller.NewBadRequest(err.Error())
	}

	if !acceptsIncomplete {
		return nil, controller.NewUnprocessableEntity()
	}

	if req.Parameters == nil {
		return nil, controller.NewBadRequest("parameters is required")
	}

	err = validateSchema(c.validator, req.Parameters)
	if err != nil {
		return nil, controller.NewBadRequest(err.Error())
	}

	payload, err := parsePayload(req.Parameters)
	if err != nil {
		return nil, controller.NewBadRequest(err.Error())
	}
	_ = payload

	// TODO implement logic

	log.Info("CreateServiceInstance kicked off creation")
	//if strings.HasSuffix(status, "_IN_PROGRESS") {
	//	return &brokerapi.CreateServiceInstanceResponse{
	//		Operation: OperationCreate,
	//		Async:     true,
	//	}, nil
	//}

	return &brokerapi.CreateServiceInstanceResponse{}, nil
}

func (c *brokernetesController) UpdateServiceInstance(ctx context.Context, instanceID string, acceptsIncomplete bool, req *brokerapi.UpdateServiceInstanceRequest) (*brokerapi.UpdateServiceInstanceResponse, error) {
	log := ctx.Value("log").(*zap.Logger)
	log = log.With(zappers.InstanceID(instanceID))
	log.Info("UpdateServiceInstance called")

	serviceId := req.ServiceID
	var planId string
	if req.PlanID == nil {
		planId = PlanId
	} else {
		planId = *req.PlanID
	}

	err := validatePlan(serviceId, planId)
	if err != nil {
		return nil, controller.NewBadRequest(err.Error())
	}

	if !acceptsIncomplete {
		return nil, controller.NewUnprocessableEntity()
	}

	if req.Parameters == nil {
		return nil, controller.NewBadRequest("parameters is required")
	}

	err = validateSchema(c.validator, req.Parameters)
	if err != nil {
		return nil, controller.NewBadRequest(err.Error())
	}

	payload, err := parsePayload(req.Parameters)
	if err != nil {
		return nil, controller.NewBadRequest(err.Error())
	}

	_ = payload
	// TODO update

	var isUpdated bool

	if isUpdated {
		return &brokerapi.UpdateServiceInstanceResponse{
			Async:     true,
			Operation: OperationUpdate,
		}, nil
	}
	return &brokerapi.UpdateServiceInstanceResponse{}, nil
}

func (c *brokernetesController) RemoveServiceInstance(ctx context.Context, instanceID, serviceID, planID string, acceptsIncomplete bool) (*brokerapi.DeleteServiceInstanceResponse, error) {
	log := ctx.Value("log").(*zap.Logger)
	log = log.With(zappers.InstanceID(instanceID))
	log.Info("RemoveServiceInstance called", zappers.ServiceID(serviceID), zappers.PlanID(planID))

	err := validatePlan(serviceID, planID)
	if err != nil {
		return nil, controller.NewBadRequest(err.Error())
	}

	if !acceptsIncomplete {
		return nil, controller.NewUnprocessableEntity()
	}

	// TODO

	log.Info("RemoveServiceInstance kicked off a deletion")

	return &brokerapi.DeleteServiceInstanceResponse{
		Operation: OperationDelete,
		Async:     true,
	}, nil
}

func (c *brokernetesController) Bind(ctx context.Context, instanceID, bindingID string, req *brokerapi.BindingRequest) (*brokerapi.CreateServiceBindingResponse, error) {
	log := ctx.Value("log").(*zap.Logger)
	log = log.With(zappers.InstanceID(instanceID))
	log.Info("Bind called", zappers.BindingID(bindingID))

	err := validatePlan(req.ServiceID, req.PlanID)
	if err != nil {
		return nil, controller.NewBadRequest(err.Error())
	}

	// TODO

	credentials := make(map[string]interface{})

	log.Info("Bind successful")
	return &brokerapi.CreateServiceBindingResponse{
		Credentials: brokerapi.Credential(credentials),
	}, nil
}

func (c *brokernetesController) UnBind(ctx context.Context, instanceID, bindingID, serviceID, planID string) error {
	log := ctx.Value("log").(*zap.Logger)
	log = log.With(zappers.InstanceID(instanceID))
	log.Info("UnBind called", zappers.BindingID(bindingID), zappers.ServiceID(serviceID), zappers.PlanID(planID))
	err := validatePlan(serviceID, planID)
	if err != nil {
		return controller.NewBadRequest(err.Error())
	}
	return err
}

func (c *brokernetesController) GetServiceInstanceStatus(ctx context.Context, instanceID, serviceID, planID, operation string) (*brokerapi.GetServiceInstanceStatusResponse, error) {
	log := ctx.Value("log").(*zap.Logger)
	log = log.With(zappers.InstanceID(instanceID))
	log.Info(
		"GetServiceInstanceStatus called",
		zappers.ServiceID(serviceID),
		zappers.PlanID(planID),
		zappers.Operation(operation))

	err := validatePlan(serviceID, planID)
	if err != nil {
		return nil, controller.NewBadRequest(err.Error())
	}

	if operation == "" {
		return nil, controller.NewBadRequest("operation must not be an empty string for this provider")
	}

	var status string
	var description string
	// TODO

	return &brokerapi.GetServiceInstanceStatusResponse{
		State:       status,
		Description: description,
	}, nil
}

func parsePayload(parameters json.RawMessage) (map[string]interface{}, error) {
	parametersMap := make(map[string]interface{})
	err := json.Unmarshal(parameters, &parametersMap)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling as json object failed")
	}
	return parametersMap, nil
}
