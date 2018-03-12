package broker

import "github.com/nilebox/broker-server/pkg/stateful/task"

var _ task.Broker = &crdBroker{}
