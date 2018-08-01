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
	v1beta1 "github.com/kubeless/http-trigger/pkg/apis/kubeless/v1beta1"
	scheme "github.com/kubeless/http-trigger/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// HTTPTriggersGetter has a method to return a HTTPTriggerInterface.
// A group's client should implement this interface.
type HTTPTriggersGetter interface {
	HTTPTriggers(namespace string) HTTPTriggerInterface
}

// HTTPTriggerInterface has methods to work with HTTPTrigger resources.
type HTTPTriggerInterface interface {
	Create(*v1beta1.HTTPTrigger) (*v1beta1.HTTPTrigger, error)
	Update(*v1beta1.HTTPTrigger) (*v1beta1.HTTPTrigger, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.HTTPTrigger, error)
	List(opts v1.ListOptions) (*v1beta1.HTTPTriggerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.HTTPTrigger, err error)
	HTTPTriggerExpansion
}

// hTTPTriggers implements HTTPTriggerInterface
type hTTPTriggers struct {
	client rest.Interface
	ns     string
}

// newHTTPTriggers returns a HTTPTriggers
func newHTTPTriggers(c *KubelessV1beta1Client, namespace string) *hTTPTriggers {
	return &hTTPTriggers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the hTTPTrigger, and returns the corresponding hTTPTrigger object, and an error if there is any.
func (c *hTTPTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.HTTPTrigger, err error) {
	result = &v1beta1.HTTPTrigger{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("httptriggers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of HTTPTriggers that match those selectors.
func (c *hTTPTriggers) List(opts v1.ListOptions) (result *v1beta1.HTTPTriggerList, err error) {
	result = &v1beta1.HTTPTriggerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("httptriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested hTTPTriggers.
func (c *hTTPTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("httptriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a hTTPTrigger and creates it.  Returns the server's representation of the hTTPTrigger, and an error, if there is any.
func (c *hTTPTriggers) Create(hTTPTrigger *v1beta1.HTTPTrigger) (result *v1beta1.HTTPTrigger, err error) {
	result = &v1beta1.HTTPTrigger{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("httptriggers").
		Body(hTTPTrigger).
		Do().
		Into(result)
	return
}

// Update takes the representation of a hTTPTrigger and updates it. Returns the server's representation of the hTTPTrigger, and an error, if there is any.
func (c *hTTPTriggers) Update(hTTPTrigger *v1beta1.HTTPTrigger) (result *v1beta1.HTTPTrigger, err error) {
	result = &v1beta1.HTTPTrigger{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("httptriggers").
		Name(hTTPTrigger.Name).
		Body(hTTPTrigger).
		Do().
		Into(result)
	return
}

// Delete takes name of the hTTPTrigger and deletes it. Returns an error if one occurs.
func (c *hTTPTriggers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("httptriggers").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *hTTPTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("httptriggers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched hTTPTrigger.
func (c *hTTPTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.HTTPTrigger, err error) {
	result = &v1beta1.HTTPTrigger{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("httptriggers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
