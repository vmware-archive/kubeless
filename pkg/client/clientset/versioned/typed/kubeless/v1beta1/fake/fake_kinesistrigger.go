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
package fake

import (
	v1beta1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKinesisTriggers implements KinesisTriggerInterface
type FakeKinesisTriggers struct {
	Fake *FakeKubelessV1beta1
	ns   string
}

var kinesistriggersResource = schema.GroupVersionResource{Group: "kubeless.io", Version: "v1beta1", Resource: "kinesistriggers"}

var kinesistriggersKind = schema.GroupVersionKind{Group: "kubeless.io", Version: "v1beta1", Kind: "KinesisTrigger"}

// Get takes name of the kinesisTrigger, and returns the corresponding kinesisTrigger object, and an error if there is any.
func (c *FakeKinesisTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.KinesisTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kinesistriggersResource, c.ns, name), &v1beta1.KinesisTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KinesisTrigger), err
}

// List takes label and field selectors, and returns the list of KinesisTriggers that match those selectors.
func (c *FakeKinesisTriggers) List(opts v1.ListOptions) (result *v1beta1.KinesisTriggerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kinesistriggersResource, kinesistriggersKind, c.ns, opts), &v1beta1.KinesisTriggerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.KinesisTriggerList{}
	for _, item := range obj.(*v1beta1.KinesisTriggerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kinesisTriggers.
func (c *FakeKinesisTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kinesistriggersResource, c.ns, opts))

}

// Create takes the representation of a kinesisTrigger and creates it.  Returns the server's representation of the kinesisTrigger, and an error, if there is any.
func (c *FakeKinesisTriggers) Create(kinesisTrigger *v1beta1.KinesisTrigger) (result *v1beta1.KinesisTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kinesistriggersResource, c.ns, kinesisTrigger), &v1beta1.KinesisTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KinesisTrigger), err
}

// Update takes the representation of a kinesisTrigger and updates it. Returns the server's representation of the kinesisTrigger, and an error, if there is any.
func (c *FakeKinesisTriggers) Update(kinesisTrigger *v1beta1.KinesisTrigger) (result *v1beta1.KinesisTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kinesistriggersResource, c.ns, kinesisTrigger), &v1beta1.KinesisTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KinesisTrigger), err
}

// Delete takes name of the kinesisTrigger and deletes it. Returns an error if one occurs.
func (c *FakeKinesisTriggers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(kinesistriggersResource, c.ns, name), &v1beta1.KinesisTrigger{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKinesisTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kinesistriggersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.KinesisTriggerList{})
	return err
}

// Patch applies the patch and returns the patched kinesisTrigger.
func (c *FakeKinesisTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.KinesisTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kinesistriggersResource, c.ns, name, data, subresources...), &v1beta1.KinesisTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KinesisTrigger), err
}
