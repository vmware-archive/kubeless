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
	v1beta1 "github.com/kubeless/kinesis-trigger/pkg/apis/kubeless/v1beta1"
	scheme "github.com/kubeless/kinesis-trigger/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KinesisTriggersGetter has a method to return a KinesisTriggerInterface.
// A group's client should implement this interface.
type KinesisTriggersGetter interface {
	KinesisTriggers(namespace string) KinesisTriggerInterface
}

// KinesisTriggerInterface has methods to work with KinesisTrigger resources.
type KinesisTriggerInterface interface {
	Create(*v1beta1.KinesisTrigger) (*v1beta1.KinesisTrigger, error)
	Update(*v1beta1.KinesisTrigger) (*v1beta1.KinesisTrigger, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.KinesisTrigger, error)
	List(opts v1.ListOptions) (*v1beta1.KinesisTriggerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.KinesisTrigger, err error)
	KinesisTriggerExpansion
}

// kinesisTriggers implements KinesisTriggerInterface
type kinesisTriggers struct {
	client rest.Interface
	ns     string
}

// newKinesisTriggers returns a KinesisTriggers
func newKinesisTriggers(c *KubelessV1beta1Client, namespace string) *kinesisTriggers {
	return &kinesisTriggers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kinesisTrigger, and returns the corresponding kinesisTrigger object, and an error if there is any.
func (c *kinesisTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.KinesisTrigger, err error) {
	result = &v1beta1.KinesisTrigger{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kinesistriggers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KinesisTriggers that match those selectors.
func (c *kinesisTriggers) List(opts v1.ListOptions) (result *v1beta1.KinesisTriggerList, err error) {
	result = &v1beta1.KinesisTriggerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kinesistriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kinesisTriggers.
func (c *kinesisTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kinesistriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a kinesisTrigger and creates it.  Returns the server's representation of the kinesisTrigger, and an error, if there is any.
func (c *kinesisTriggers) Create(kinesisTrigger *v1beta1.KinesisTrigger) (result *v1beta1.KinesisTrigger, err error) {
	result = &v1beta1.KinesisTrigger{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kinesistriggers").
		Body(kinesisTrigger).
		Do().
		Into(result)
	return
}

// Update takes the representation of a kinesisTrigger and updates it. Returns the server's representation of the kinesisTrigger, and an error, if there is any.
func (c *kinesisTriggers) Update(kinesisTrigger *v1beta1.KinesisTrigger) (result *v1beta1.KinesisTrigger, err error) {
	result = &v1beta1.KinesisTrigger{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kinesistriggers").
		Name(kinesisTrigger.Name).
		Body(kinesisTrigger).
		Do().
		Into(result)
	return
}

// Delete takes name of the kinesisTrigger and deletes it. Returns an error if one occurs.
func (c *kinesisTriggers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kinesistriggers").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kinesisTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kinesistriggers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched kinesisTrigger.
func (c *kinesisTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.KinesisTrigger, err error) {
	result = &v1beta1.KinesisTrigger{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kinesistriggers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
