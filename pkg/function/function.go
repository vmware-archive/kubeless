/*
Copyright 2016 Skippbox, Ltd.

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

package function

import (
	"sync"

	"github.com/kubeless/kubeless/pkg/spec"
	"github.com/kubeless/kubeless/pkg/utils"

	"k8s.io/client-go/kubernetes"
)

type functionEventType string

type functionEvent struct {
	typ  functionEventType
	spec spec.FunctionSpec
}

// New creates the custom function object
func New(c kubernetes.Interface, name, ns string, spec *spec.FunctionSpec, wg *sync.WaitGroup) error {
	return new(c, name, ns, spec, wg)
}

// Delete removes the custom function object
func Delete(c kubernetes.Interface, name, ns string, wg *sync.WaitGroup) error {
	return delete(c, name, ns, wg)
}

// Update apply changes to the custom function object
func Update(c kubernetes.Interface, name, ns string, spec *spec.FunctionSpec, wg *sync.WaitGroup) error {
	return update(c, name, ns, spec, wg)
}

func new(kclient kubernetes.Interface, name, ns string, spec *spec.FunctionSpec, wg *sync.WaitGroup) error {
	err := utils.CreateK8sResources(ns, name, spec, kclient)
	if err != nil {
		return err
	}

	wg.Add(1)

	return nil
}

func delete(kclient kubernetes.Interface, name, ns string, wg *sync.WaitGroup) error {
	err := utils.DeleteK8sResources(ns, name, kclient)
	wg.Add(1)
	return err
}

func update(kclient kubernetes.Interface, name, ns string, spec *spec.FunctionSpec, wg *sync.WaitGroup) error {
	err := utils.UpdateK8sResources(kclient, name, ns, spec)
	if err != nil {
		return err
	}

	wg.Add(1)
	return nil
}
