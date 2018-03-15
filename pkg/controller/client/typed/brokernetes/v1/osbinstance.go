// Generated file, do not modify manually!

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/nilebox/brokernetes/pkg/apis/brokernetes/v1"
	scheme "github.com/nilebox/brokernetes/pkg/controller/client/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// OSBInstancesGetter has a method to return a OSBInstanceInterface.
// A group's client should implement this interface.
type OSBInstancesGetter interface {
	OSBInstances(namespace string) OSBInstanceInterface
}

// OSBInstanceInterface has methods to work with OSBInstance resources.
type OSBInstanceInterface interface {
	Create(*v1.OSBInstance) (*v1.OSBInstance, error)
	Update(*v1.OSBInstance) (*v1.OSBInstance, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.OSBInstance, error)
	List(opts meta_v1.ListOptions) (*v1.OSBInstanceList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.OSBInstance, err error)
	OSBInstanceExpansion
}

// oSBInstances implements OSBInstanceInterface
type oSBInstances struct {
	client rest.Interface
	ns     string
}

// newOSBInstances returns a OSBInstances
func newOSBInstances(c *BrokernetesV1Client, namespace string) *oSBInstances {
	return &oSBInstances{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the oSBInstance, and returns the corresponding oSBInstance object, and an error if there is any.
func (c *oSBInstances) Get(name string, options meta_v1.GetOptions) (result *v1.OSBInstance, err error) {
	result = &v1.OSBInstance{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("osbinstances").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of OSBInstances that match those selectors.
func (c *oSBInstances) List(opts meta_v1.ListOptions) (result *v1.OSBInstanceList, err error) {
	result = &v1.OSBInstanceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("osbinstances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested oSBInstances.
func (c *oSBInstances) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("osbinstances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a oSBInstance and creates it.  Returns the server's representation of the oSBInstance, and an error, if there is any.
func (c *oSBInstances) Create(oSBInstance *v1.OSBInstance) (result *v1.OSBInstance, err error) {
	result = &v1.OSBInstance{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("osbinstances").
		Body(oSBInstance).
		Do().
		Into(result)
	return
}

// Update takes the representation of a oSBInstance and updates it. Returns the server's representation of the oSBInstance, and an error, if there is any.
func (c *oSBInstances) Update(oSBInstance *v1.OSBInstance) (result *v1.OSBInstance, err error) {
	result = &v1.OSBInstance{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("osbinstances").
		Name(oSBInstance.Name).
		Body(oSBInstance).
		Do().
		Into(result)
	return
}

// Delete takes name of the oSBInstance and deletes it. Returns an error if one occurs.
func (c *oSBInstances) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("osbinstances").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *oSBInstances) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("osbinstances").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched oSBInstance.
func (c *oSBInstances) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.OSBInstance, err error) {
	result = &v1.OSBInstance{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("osbinstances").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
