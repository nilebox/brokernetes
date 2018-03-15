package controller

import (
	"context"
	"log"
	"time"

	"github.com/nilebox/brokernetes/pkg/apis/brokernetes/v1"
	client_v1 "github.com/nilebox/brokernetes/pkg/controller/client/typed/brokernetes/v1"
	"github.com/nilebox/brokernetes/pkg/controller/manager"

	"github.com/ash2k/stager/wait"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// Work queue deduplicates scheduled keys. This is the period it waits for duplicate keys before letting the work
	// to be dequeued.
	workDeduplicationPeriod = 50 * time.Millisecond
)

type Controller struct {
	osbInstanceInformer cache.SharedIndexInformer
	queue               workqueue.RateLimitingInterface
	workers             int
	manager             *manager.Manager
}

func OsbInstanceInformer(client client_v1.OSBInstancesGetter, namespace string, resyncPeriod time.Duration) cache.SharedIndexInformer {
	instances := client.OSBInstances(namespace)
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return instances.List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return instances.Watch(options)
			},
		},
		&v1.OSBInstance{},
		resyncPeriod,
		cache.Indexers{})
}

func NewController(osbInstanceInformer cache.SharedIndexInformer, workers int) *Controller {
	c := &Controller{
		osbInstanceInformer: osbInstanceInformer,
		queue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "osbInstance"),
		workers:             workers,
	}
	osbInstanceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onOSBInstanceAdd,
		UpdateFunc: c.onOSBInstanceUpdate,
		DeleteFunc: c.onOSBInstanceDelete,
	})
	osbInstanceInformer.AddIndexers(cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	})

	return c
}

// Run begins watching and syncing.
func (c *Controller) Run(ctx context.Context) {
	var wg wait.Group
	defer wg.Wait()
	defer c.queue.ShutDown()

	log.Print("Starting OSBInstance controller")
	defer log.Print("Shutting down OSBInstance controller")

	if !cache.WaitForCacheSync(ctx.Done(), c.osbInstanceInformer.HasSynced) {
		return
	}

	log.Print("Finished syncing caches")

	for i := 0; i < c.workers; i++ {
		wg.StartWithContext(ctx, c.worker)
	}

	<-ctx.Done()
}

func (c *Controller) enqueue(state *v1.OSBInstance) {
	key, err := cache.MetaNamespaceKeyFunc(state)
	if err != nil {
		log.Printf("Couldn't get key for OSBInstance %+v: %v", state, err)
		return
	}
	c.enqueueKey(key)
}

func (c *Controller) enqueueKey(key string) {
	c.queue.AddAfter(key, workDeduplicationPeriod)
}
