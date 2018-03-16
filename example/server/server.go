package server

import (
	"context"

	"go.uber.org/zap"
	"github.com/nilebox/broker-server/pkg/stateful"
	"github.com/nilebox/broker-server/pkg/stateful/task"
	"github.com/nilebox/broker-server/pkg/api"
	"github.com/nilebox/broker-server/pkg/stateful/storage"
	"github.com/nilebox/broker-server/pkg/server"
)

type ExampleServer struct {
	Addr string
}

func (b *ExampleServer) Run(ctx context.Context, log *zap.Logger, catalog *api.Catalog, broker task.Broker, storage storage.Storage) (returnErr error) {
	// Run a REST server
	controller := stateful.NewStatefulController(ctx, catalog, storage)
	return server.Run(ctx, b.Addr, controller)
}
