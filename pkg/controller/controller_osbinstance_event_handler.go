package controller

import (
	"log"

	"github.com/nilebox/brokernetes/pkg/apis/brokernetes/v1"

	"k8s.io/client-go/tools/cache"
)

func (c *Controller) onOSBInstanceAdd(obj interface{}) {
	c.enqueue(obj.(*v1.OSBInstance))
}

func (c *Controller) onOSBInstanceUpdate(oldObj, newObj interface{}) {
	c.enqueue(newObj.(*v1.OSBInstance))
}

func (c *Controller) onOSBInstanceDelete(obj interface{}) {
	osbInstance, ok := obj.(*v1.OSBInstance)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			log.Printf("Couldn't get OSBInstance from tombstone %#v", obj)
			return
		}
		c.enqueueKey(tombstone.Key)
		return
	}
	c.enqueue(osbInstance)
}
