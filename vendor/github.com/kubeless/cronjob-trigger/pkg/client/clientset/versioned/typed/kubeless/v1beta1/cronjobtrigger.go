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
	v1beta1 "github.com/kubeless/cronjob-trigger/pkg/apis/kubeless/v1beta1"
	scheme "github.com/kubeless/cronjob-trigger/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CronJobTriggersGetter has a method to return a CronJobTriggerInterface.
// A group's client should implement this interface.
type CronJobTriggersGetter interface {
	CronJobTriggers(namespace string) CronJobTriggerInterface
}

// CronJobTriggerInterface has methods to work with CronJobTrigger resources.
type CronJobTriggerInterface interface {
	Create(*v1beta1.CronJobTrigger) (*v1beta1.CronJobTrigger, error)
	Update(*v1beta1.CronJobTrigger) (*v1beta1.CronJobTrigger, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.CronJobTrigger, error)
	List(opts v1.ListOptions) (*v1beta1.CronJobTriggerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.CronJobTrigger, err error)
	CronJobTriggerExpansion
}

// cronJobTriggers implements CronJobTriggerInterface
type cronJobTriggers struct {
	client rest.Interface
	ns     string
}

// newCronJobTriggers returns a CronJobTriggers
func newCronJobTriggers(c *KubelessV1beta1Client, namespace string) *cronJobTriggers {
	return &cronJobTriggers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cronJobTrigger, and returns the corresponding cronJobTrigger object, and an error if there is any.
func (c *cronJobTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.CronJobTrigger, err error) {
	result = &v1beta1.CronJobTrigger{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cronjobtriggers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CronJobTriggers that match those selectors.
func (c *cronJobTriggers) List(opts v1.ListOptions) (result *v1beta1.CronJobTriggerList, err error) {
	result = &v1beta1.CronJobTriggerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cronjobtriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cronJobTriggers.
func (c *cronJobTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cronjobtriggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a cronJobTrigger and creates it.  Returns the server's representation of the cronJobTrigger, and an error, if there is any.
func (c *cronJobTriggers) Create(cronJobTrigger *v1beta1.CronJobTrigger) (result *v1beta1.CronJobTrigger, err error) {
	result = &v1beta1.CronJobTrigger{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cronjobtriggers").
		Body(cronJobTrigger).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cronJobTrigger and updates it. Returns the server's representation of the cronJobTrigger, and an error, if there is any.
func (c *cronJobTriggers) Update(cronJobTrigger *v1beta1.CronJobTrigger) (result *v1beta1.CronJobTrigger, err error) {
	result = &v1beta1.CronJobTrigger{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cronjobtriggers").
		Name(cronJobTrigger.Name).
		Body(cronJobTrigger).
		Do().
		Into(result)
	return
}

// Delete takes name of the cronJobTrigger and deletes it. Returns an error if one occurs.
func (c *cronJobTriggers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cronjobtriggers").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cronJobTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cronjobtriggers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cronJobTrigger.
func (c *cronJobTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.CronJobTrigger, err error) {
	result = &v1beta1.CronJobTrigger{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cronjobtriggers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
