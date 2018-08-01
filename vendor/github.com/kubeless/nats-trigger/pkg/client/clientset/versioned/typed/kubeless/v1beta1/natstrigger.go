/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1beta1

import (
	v1beta1 "github.com/kubeless/nats-trigger/pkg/apis/kubeless/v1beta1"
	scheme "github.com/kubeless/nats-trigger/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// NATSTriggersGetter has a method to return a NATSTriggerInterface.
// A group's client should implement this interface.
type NATSTriggersGetter interface {
	NATSTriggers(namespace string) NATSTriggerInterface
}

// NATSTriggerInterface has methods to work with NATSTrigger resources.
type NATSTriggerInterface interface {
	Create(*v1beta1.NATSTrigger) (*v1beta1.NATSTrigger, error)
	Update(*v1beta1.NATSTrigger) (*v1beta1.NATSTrigger, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.NATSTrigger, error)
	List(opts v1.ListOptions) (*v1beta1.NATSTriggerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.NATSTrigger, err error)
	NATSTriggerExpansion
}

// nATSTriggers implements NATSTriggerInterface
type nATSTriggers struct {
	client rest.Interface
	ns     string
}

// newNATSTriggers returns a NATSTriggers
func newNATSTriggers(c *KubelessV1beta1Client, namespace string) *nATSTriggers {
	return &nATSTriggers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the nATSTrigger, and returns the corresponding nATSTrigger object, and an error if there is any.
func (c *nATSTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.NATSTrigger, err error) {
	result = &v1beta1.NATSTrigger{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("natstriggers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NATSTriggers that match those selectors.
func (c *nATSTriggers) List(opts v1.ListOptions) (result *v1beta1.NATSTriggerList, err error) {
	result = &v1beta1.NATSTriggerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("natstriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested nATSTriggers.
func (c *nATSTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("natstriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a nATSTrigger and creates it.  Returns the server's representation of the nATSTrigger, and an error, if there is any.
func (c *nATSTriggers) Create(nATSTrigger *v1beta1.NATSTrigger) (result *v1beta1.NATSTrigger, err error) {
	result = &v1beta1.NATSTrigger{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("natstriggers").
		Body(nATSTrigger).
		Do().
		Into(result)
	return
}

// Update takes the representation of a nATSTrigger and updates it. Returns the server's representation of the nATSTrigger, and an error, if there is any.
func (c *nATSTriggers) Update(nATSTrigger *v1beta1.NATSTrigger) (result *v1beta1.NATSTrigger, err error) {
	result = &v1beta1.NATSTrigger{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("natstriggers").
		Name(nATSTrigger.Name).
		Body(nATSTrigger).
		Do().
		Into(result)
	return
}

// Delete takes name of the nATSTrigger and deletes it. Returns an error if one occurs.
func (c *nATSTriggers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("natstriggers").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *nATSTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("natstriggers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched nATSTrigger.
func (c *nATSTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.NATSTrigger, err error) {
	result = &v1beta1.NATSTrigger{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("natstriggers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
