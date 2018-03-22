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

// FakeKafkaTriggers implements KafkaTriggerInterface
type FakeKafkaTriggers struct {
	Fake *FakeKubelessV1beta1
	ns   string
}

var kafkatriggersResource = schema.GroupVersionResource{Group: "kubeless.io", Version: "v1beta1", Resource: "kafkatriggers"}

var kafkatriggersKind = schema.GroupVersionKind{Group: "kubeless.io", Version: "v1beta1", Kind: "KafkaTrigger"}

// Get takes name of the kafkaTrigger, and returns the corresponding kafkaTrigger object, and an error if there is any.
func (c *FakeKafkaTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.KafkaTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kafkatriggersResource, c.ns, name), &v1beta1.KafkaTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KafkaTrigger), err
}

// List takes label and field selectors, and returns the list of KafkaTriggers that match those selectors.
func (c *FakeKafkaTriggers) List(opts v1.ListOptions) (result *v1beta1.KafkaTriggerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kafkatriggersResource, kafkatriggersKind, c.ns, opts), &v1beta1.KafkaTriggerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.KafkaTriggerList{}
	for _, item := range obj.(*v1beta1.KafkaTriggerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kafkaTriggers.
func (c *FakeKafkaTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kafkatriggersResource, c.ns, opts))

}

// Create takes the representation of a kafkaTrigger and creates it.  Returns the server's representation of the kafkaTrigger, and an error, if there is any.
func (c *FakeKafkaTriggers) Create(kafkaTrigger *v1beta1.KafkaTrigger) (result *v1beta1.KafkaTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kafkatriggersResource, c.ns, kafkaTrigger), &v1beta1.KafkaTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KafkaTrigger), err
}

// Update takes the representation of a kafkaTrigger and updates it. Returns the server's representation of the kafkaTrigger, and an error, if there is any.
func (c *FakeKafkaTriggers) Update(kafkaTrigger *v1beta1.KafkaTrigger) (result *v1beta1.KafkaTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kafkatriggersResource, c.ns, kafkaTrigger), &v1beta1.KafkaTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KafkaTrigger), err
}

// Delete takes name of the kafkaTrigger and deletes it. Returns an error if one occurs.
func (c *FakeKafkaTriggers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(kafkatriggersResource, c.ns, name), &v1beta1.KafkaTrigger{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKafkaTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kafkatriggersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.KafkaTriggerList{})
	return err
}

// Patch applies the patch and returns the patched kafkaTrigger.
func (c *FakeKafkaTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.KafkaTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kafkatriggersResource, c.ns, name, data, subresources...), &v1beta1.KafkaTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KafkaTrigger), err
}
