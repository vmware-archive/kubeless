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

// FakeHTTPTriggers implements HTTPTriggerInterface
type FakeHTTPTriggers struct {
	Fake *FakeKubelessV1beta1
	ns   string
}

var httptriggersResource = schema.GroupVersionResource{Group: "kubeless.io", Version: "v1beta1", Resource: "httptriggers"}

var httptriggersKind = schema.GroupVersionKind{Group: "kubeless.io", Version: "v1beta1", Kind: "HTTPTrigger"}

// Get takes name of the hTTPTrigger, and returns the corresponding hTTPTrigger object, and an error if there is any.
func (c *FakeHTTPTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.HTTPTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(httptriggersResource, c.ns, name), &v1beta1.HTTPTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.HTTPTrigger), err
}

// List takes label and field selectors, and returns the list of HTTPTriggers that match those selectors.
func (c *FakeHTTPTriggers) List(opts v1.ListOptions) (result *v1beta1.HTTPTriggerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(httptriggersResource, httptriggersKind, c.ns, opts), &v1beta1.HTTPTriggerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.HTTPTriggerList{}
	for _, item := range obj.(*v1beta1.HTTPTriggerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested hTTPTriggers.
func (c *FakeHTTPTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(httptriggersResource, c.ns, opts))

}

// Create takes the representation of a hTTPTrigger and creates it.  Returns the server's representation of the hTTPTrigger, and an error, if there is any.
func (c *FakeHTTPTriggers) Create(hTTPTrigger *v1beta1.HTTPTrigger) (result *v1beta1.HTTPTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(httptriggersResource, c.ns, hTTPTrigger), &v1beta1.HTTPTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.HTTPTrigger), err
}

// Update takes the representation of a hTTPTrigger and updates it. Returns the server's representation of the hTTPTrigger, and an error, if there is any.
func (c *FakeHTTPTriggers) Update(hTTPTrigger *v1beta1.HTTPTrigger) (result *v1beta1.HTTPTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(httptriggersResource, c.ns, hTTPTrigger), &v1beta1.HTTPTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.HTTPTrigger), err
}

// Delete takes name of the hTTPTrigger and deletes it. Returns an error if one occurs.
func (c *FakeHTTPTriggers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(httptriggersResource, c.ns, name), &v1beta1.HTTPTrigger{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeHTTPTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(httptriggersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.HTTPTriggerList{})
	return err
}

// Patch applies the patch and returns the patched hTTPTrigger.
func (c *FakeHTTPTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.HTTPTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(httptriggersResource, c.ns, name, data, subresources...), &v1beta1.HTTPTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.HTTPTrigger), err
}
