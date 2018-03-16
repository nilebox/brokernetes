package main

import (
	"context"
	"time"
	"os"

	"github.com/nilebox/brokernetes/pkg/client"
	"github.com/nilebox/brokernetes/pkg/controller"

	osbinstance_client "github.com/nilebox/brokernetes/pkg/controller/client/typed/brokernetes/v1"
	clientset_client "github.com/nilebox/brokernetes/pkg/controller/client"

	"github.com/ash2k/stager"
	manager "github.com/nilebox/brokernetes/pkg/controller/manager"
	broker "github.com/nilebox/brokernetes/example/broker"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	storage "github.com/nilebox/brokernetes/pkg/storage"
	server2 "github.com/nilebox/brokernetes/example/server"
	"os/signal"
	"syscall"
)

const (
	defaultAddr = ":8080"

)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	cancelOnInterrupt(ctx, cancelFunc)


	// Create kubernetes controller and run it
	restConfig, err := client.ConfigFromEnv()
	if err != nil {
		panic(err)
	}

	clientset := clientset_client.NewForConfigOrDie(restConfig)
	client, err := osbinstance_client.NewForConfig(restConfig)

	if err != nil {
		panic("Could not create osbinstance client")
	}

	// Create informer
	log := initializeLogger()
	defer log.Sync()

	exampleBroker, err := broker.NewExampleBroker(log)
	if err != nil {
		panic("Couldn't create a broker!")
	}

	namespace := "example_brokernetes_namespace"
	storage := storage.NewCrdStorage(client, namespace)
	manager := manager.NewManager(client.OSBInstances(namespace), exampleBroker)
	informer := controller.OsbInstanceInformer(clientset.BrokernetesV1(), namespace, time.Minute)
	c := controller.NewController(informer, 2, manager)

	// Run the informer and the controller

	stgr := stager.New()
	defer stgr.Shutdown()
	stage := stgr.NextStage()

	stage.StartWithChannel(informer.Run)

	c.Run(context.TODO())

	server := server2.ExampleServer{defaultAddr}

	server.Run(ctx, log, broker.Catalog(), exampleBroker, storage)
}

func initializeLogger() *zap.Logger {
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


// cancelOnInterrupt calls f when os.Interrupt or SIGTERM is received.
// It ignores subsequent interrupts on purpose - program should exit correctly after the first signal.
func cancelOnInterrupt(ctx context.Context, f context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-c:
			f()
		}
	}()
}
