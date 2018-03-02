package app

import (
	"context"
	"time"

	"github.com/nilebox/brokernetes/pkg/broker/brokerserver"
	"github.com/nilebox/brokernetes/pkg/brokernetes"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type BrokernetesBroker struct {
	CacheTime  time.Duration
	ExpireTime time.Duration

	Addr string
}

func (c *BrokernetesBroker) Run(ctx context.Context) (returnErr error) {
	log := ctx.Value("log").(*zap.Logger)

	controller, err := brokernetes.CreateController()
	if err != nil {
		return err
	}

	return brokerserver.Run(ctx, c.Addr, controller)
}
