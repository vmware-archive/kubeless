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

// FakeNATSTriggers implements NATSTriggerInterface
type FakeNATSTriggers struct {
	Fake *FakeKubelessV1beta1
	ns   string
}

var natstriggersResource = schema.GroupVersionResource{Group: "kubeless.io", Version: "v1beta1", Resource: "natstriggers"}

var natstriggersKind = schema.GroupVersionKind{Group: "kubeless.io", Version: "v1beta1", Kind: "NATSTrigger"}

// Get takes name of the nATSTrigger, and returns the corresponding nATSTrigger object, and an error if there is any.
func (c *FakeNATSTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.NATSTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(natstriggersResource, c.ns, name), &v1beta1.NATSTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.NATSTrigger), err
}

// List takes label and field selectors, and returns the list of NATSTriggers that match those selectors.
func (c *FakeNATSTriggers) List(opts v1.ListOptions) (result *v1beta1.NATSTriggerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(natstriggersResource, natstriggersKind, c.ns, opts), &v1beta1.NATSTriggerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.NATSTriggerList{}
	for _, item := range obj.(*v1beta1.NATSTriggerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested nATSTriggers.
func (c *FakeNATSTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(natstriggersResource, c.ns, opts))

}

// Create takes the representation of a nATSTrigger and creates it.  Returns the server's representation of the nATSTrigger, and an error, if there is any.
func (c *FakeNATSTriggers) Create(nATSTrigger *v1beta1.NATSTrigger) (result *v1beta1.NATSTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(natstriggersResource, c.ns, nATSTrigger), &v1beta1.NATSTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.NATSTrigger), err
}

// Update takes the representation of a nATSTrigger and updates it. Returns the server's representation of the nATSTrigger, and an error, if there is any.
func (c *FakeNATSTriggers) Update(nATSTrigger *v1beta1.NATSTrigger) (result *v1beta1.NATSTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(natstriggersResource, c.ns, nATSTrigger), &v1beta1.NATSTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.NATSTrigger), err
}

// Delete takes name of the nATSTrigger and deletes it. Returns an error if one occurs.
func (c *FakeNATSTriggers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(natstriggersResource, c.ns, name), &v1beta1.NATSTrigger{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeNATSTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(natstriggersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.NATSTriggerList{})
	return err
}

// Patch applies the patch and returns the patched nATSTrigger.
func (c *FakeNATSTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.NATSTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(natstriggersResource, c.ns, name, data, subresources...), &v1beta1.NATSTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.NATSTrigger), err
}
