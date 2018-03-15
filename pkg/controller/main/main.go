package main

import (
	"context"
	"time"

	"github.com/nilebox/brokernetes/pkg/client"
	"github.com/nilebox/brokernetes/pkg/controller"
	osbinstance_client "github.com/nilebox/brokernetes/pkg/controller/client"

	"github.com/ash2k/stager"
)

func main() {
	// Create kubernetes controller and run it
	restConfig, err := client.ConfigFromEnv()
	if err != nil {
		panic(err)
	}

	clientset := osbinstance_client.NewForConfigOrDie(restConfig)

	// Create informer
	informer := controller.OsbInstanceInformer(clientset.BrokernetesV1(), "ivor", time.Minute)
	c := controller.NewController(informer, 2)

	// Run the informer and the controller

	stgr := stager.New()
	defer stgr.Shutdown()
	stage := stgr.NextStage()

	stage.StartWithChannel(informer.Run)

	c.Run(context.TODO())
}
