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
	v1beta1 "github.com/kubeless/kafka-trigger/pkg/apis/kubeless/v1beta1"
	scheme "github.com/kubeless/kafka-trigger/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KafkaTriggersGetter has a method to return a KafkaTriggerInterface.
// A group's client should implement this interface.
type KafkaTriggersGetter interface {
	KafkaTriggers(namespace string) KafkaTriggerInterface
}

// KafkaTriggerInterface has methods to work with KafkaTrigger resources.
type KafkaTriggerInterface interface {
	Create(*v1beta1.KafkaTrigger) (*v1beta1.KafkaTrigger, error)
	Update(*v1beta1.KafkaTrigger) (*v1beta1.KafkaTrigger, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.KafkaTrigger, error)
	List(opts v1.ListOptions) (*v1beta1.KafkaTriggerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.KafkaTrigger, err error)
	KafkaTriggerExpansion
}

// kafkaTriggers implements KafkaTriggerInterface
type kafkaTriggers struct {
	client rest.Interface
	ns     string
}

// newKafkaTriggers returns a KafkaTriggers
func newKafkaTriggers(c *KubelessV1beta1Client, namespace string) *kafkaTriggers {
	return &kafkaTriggers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kafkaTrigger, and returns the corresponding kafkaTrigger object, and an error if there is any.
func (c *kafkaTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.KafkaTrigger, err error) {
	result = &v1beta1.KafkaTrigger{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kafkatriggers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KafkaTriggers that match those selectors.
func (c *kafkaTriggers) List(opts v1.ListOptions) (result *v1beta1.KafkaTriggerList, err error) {
	result = &v1beta1.KafkaTriggerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kafkatriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kafkaTriggers.
func (c *kafkaTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kafkatriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a kafkaTrigger and creates it.  Returns the server's representation of the kafkaTrigger, and an error, if there is any.
func (c *kafkaTriggers) Create(kafkaTrigger *v1beta1.KafkaTrigger) (result *v1beta1.KafkaTrigger, err error) {
	result = &v1beta1.KafkaTrigger{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kafkatriggers").
		Body(kafkaTrigger).
		Do().
		Into(result)
	return
}

// Update takes the representation of a kafkaTrigger and updates it. Returns the server's representation of the kafkaTrigger, and an error, if there is any.
func (c *kafkaTriggers) Update(kafkaTrigger *v1beta1.KafkaTrigger) (result *v1beta1.KafkaTrigger, err error) {
	result = &v1beta1.KafkaTrigger{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kafkatriggers").
		Name(kafkaTrigger.Name).
		Body(kafkaTrigger).
		Do().
		Into(result)
	return
}

// Delete takes name of the kafkaTrigger and deletes it. Returns an error if one occurs.
func (c *kafkaTriggers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kafkatriggers").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kafkaTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kafkatriggers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched kafkaTrigger.
func (c *kafkaTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.KafkaTrigger, err error) {
	result = &v1beta1.KafkaTrigger{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kafkatriggers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
