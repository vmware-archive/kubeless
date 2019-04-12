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
	v1beta1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	scheme "github.com/kubeless/kubeless/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// FunctionsGetter has a method to return a FunctionInterface.
// A group's client should implement this interface.
type FunctionsGetter interface {
	Functions(namespace string) FunctionInterface
}

// FunctionInterface has methods to work with Function resources.
type FunctionInterface interface {
	Create(*v1beta1.Function) (*v1beta1.Function, error)
	Update(*v1beta1.Function) (*v1beta1.Function, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.Function, error)
	List(opts v1.ListOptions) (*v1beta1.FunctionList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Function, err error)
	FunctionExpansion
}

// functions implements FunctionInterface
type functions struct {
	client rest.Interface
	ns     string
}

// newFunctions returns a Functions
func newFunctions(c *KubelessV1beta1Client, namespace string) *functions {
	return &functions{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the function, and returns the corresponding function object, and an error if there is any.
func (c *functions) Get(name string, options v1.GetOptions) (result *v1beta1.Function, err error) {
	result = &v1beta1.Function{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("functions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Functions that match those selectors.
func (c *functions) List(opts v1.ListOptions) (result *v1beta1.FunctionList, err error) {
	result = &v1beta1.FunctionList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("functions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested functions.
func (c *functions) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("functions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a function and creates it.  Returns the server's representation of the function, and an error, if there is any.
func (c *functions) Create(function *v1beta1.Function) (result *v1beta1.Function, err error) {
	result = &v1beta1.Function{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("functions").
		Body(function).
		Do().
		Into(result)
	return
}

// Update takes the representation of a function and updates it. Returns the server's representation of the function, and an error, if there is any.
func (c *functions) Update(function *v1beta1.Function) (result *v1beta1.Function, err error) {
	result = &v1beta1.Function{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("functions").
		Name(function.Name).
		Body(function).
		Do().
		Into(result)
	return
}

// Delete takes name of the function and deletes it. Returns an error if one occurs.
func (c *functions) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("functions").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *functions) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("functions").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched function.
func (c *functions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Function, err error) {
	result = &v1beta1.Function{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("functions").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
