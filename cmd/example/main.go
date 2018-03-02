package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nilebox/brokernetes/cmd"
	exampleApp "github.com/nilebox/brokernetes/cmd/example/app"
)

const (
	defaultAddr = ":8080"

	defaultCacheDuration  = 10 * time.Second
	defaultExpireDuration = 1 * time.Hour
)

func main() {
	if err := run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	cmd.CancelOnInterrupt(ctx, cancelFunc)

	log := initialiseLogger()
	defer log.Sync()
	ctx = context.WithValue(ctx, "log", log)

	return runWithContext(ctx)
}

func runWithContext(ctx context.Context) error {
	log := ctx.Value("log").(*zap.Logger)
	_ = log

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	addr := fs.String("addr", defaultAddr, "Address to listen on")

	cacheMs := fs.Int64("cache_ms", int64(defaultCacheDuration/time.Millisecond), "Time to cache entries for in Millis")
	expireMs := fs.Int64("expire_ms", int64(defaultExpireDuration/time.Millisecond), "Time to cache stacks unlikely to change for in Millis")

	fs.Parse(os.Args[1:]) // nolint: gas

	app := exampleApp.ExampleBroker{
		Addr: *addr,

		CacheTime:  time.Duration(*cacheMs) * time.Millisecond,
		ExpireTime: time.Duration(*expireMs) * time.Millisecond,
	}
	return app.Run(ctx)
}

func initialiseLogger() *zap.Logger {
	return zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.Lock(zapcore.AddSync(os.Stdout)),
			zap.InfoLevel,
		),
		zap.AddCaller(),
		zap.Fields(),
	)
}
