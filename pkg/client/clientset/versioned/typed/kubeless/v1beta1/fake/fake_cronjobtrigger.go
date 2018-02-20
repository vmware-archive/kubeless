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

// FakeCronJobTriggers implements CronJobTriggerInterface
type FakeCronJobTriggers struct {
	Fake *FakeKubelessV1beta1
	ns   string
}

var cronjobtriggersResource = schema.GroupVersionResource{Group: "kubeless.io", Version: "v1beta1", Resource: "cronjobtriggers"}

var cronjobtriggersKind = schema.GroupVersionKind{Group: "kubeless.io", Version: "v1beta1", Kind: "CronJobTrigger"}

// Get takes name of the cronJobTrigger, and returns the corresponding cronJobTrigger object, and an error if there is any.
func (c *FakeCronJobTriggers) Get(name string, options v1.GetOptions) (result *v1beta1.CronJobTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cronjobtriggersResource, c.ns, name), &v1beta1.CronJobTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CronJobTrigger), err
}

// List takes label and field selectors, and returns the list of CronJobTriggers that match those selectors.
func (c *FakeCronJobTriggers) List(opts v1.ListOptions) (result *v1beta1.CronJobTriggerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cronjobtriggersResource, cronjobtriggersKind, c.ns, opts), &v1beta1.CronJobTriggerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.CronJobTriggerList{}
	for _, item := range obj.(*v1beta1.CronJobTriggerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cronJobTriggers.
func (c *FakeCronJobTriggers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cronjobtriggersResource, c.ns, opts))

}

// Create takes the representation of a cronJobTrigger and creates it.  Returns the server's representation of the cronJobTrigger, and an error, if there is any.
func (c *FakeCronJobTriggers) Create(cronJobTrigger *v1beta1.CronJobTrigger) (result *v1beta1.CronJobTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cronjobtriggersResource, c.ns, cronJobTrigger), &v1beta1.CronJobTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CronJobTrigger), err
}

// Update takes the representation of a cronJobTrigger and updates it. Returns the server's representation of the cronJobTrigger, and an error, if there is any.
func (c *FakeCronJobTriggers) Update(cronJobTrigger *v1beta1.CronJobTrigger) (result *v1beta1.CronJobTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cronjobtriggersResource, c.ns, cronJobTrigger), &v1beta1.CronJobTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CronJobTrigger), err
}

// Delete takes name of the cronJobTrigger and deletes it. Returns an error if one occurs.
func (c *FakeCronJobTriggers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cronjobtriggersResource, c.ns, name), &v1beta1.CronJobTrigger{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCronJobTriggers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cronjobtriggersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.CronJobTriggerList{})
	return err
}

// Patch applies the patch and returns the patched cronJobTrigger.
func (c *FakeCronJobTriggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.CronJobTrigger, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cronjobtriggersResource, c.ns, name, data, subresources...), &v1beta1.CronJobTrigger{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CronJobTrigger), err
}
