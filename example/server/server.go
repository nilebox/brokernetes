package server

import (
	"context"

	"go.uber.org/zap"
)

type ExampleServer struct {
	Addr string
}

func (b *ExampleServer) Run(ctx context.Context) (returnErr error) {
	log := ctx.Value("log").(*zap.Logger)
	_ = log

	panic("NotImplemented")
	//c, err := broker.NewController(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//return brokerserver.Run(ctx, b.Addr, c)
}
