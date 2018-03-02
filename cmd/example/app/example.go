package app

import (
	"context"
	"time"

	"github.com/nilebox/brokernetes/pkg/broker/brokerserver"
	"github.com/nilebox/brokernetes/pkg/controller"
	"go.uber.org/zap"
)

type ExampleBroker struct {
	CacheTime  time.Duration
	ExpireTime time.Duration

	Addr string
}

func (b *ExampleBroker) Run(ctx context.Context) (returnErr error) {
	log := ctx.Value("log").(*zap.Logger)
	_ = log

	c, err := controller.NewController(ctx)
	if err != nil {
		return err
	}

	return brokerserver.Run(ctx, b.Addr, c)
}
