package zappers

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BindingID(bindingID string) zapcore.Field {
	return zap.String("binding_id", bindingID)
}

func InstanceID(instanceID string) zapcore.Field {
	return zap.String("instance_id", instanceID)
}

func Operation(operation string) zapcore.Field {
	return zap.String("operation", operation)
}

func PlanID(planID string) zapcore.Field {
	return zap.String("plan_id", planID)
}

func ServiceID(serviceID string) zapcore.Field {
	return zap.String("service_id", serviceID)
}
