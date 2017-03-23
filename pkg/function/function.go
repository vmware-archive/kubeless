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

	"github.com/Sirupsen/logrus"
	"github.com/bitnami/kubeless/pkg/spec"
	"github.com/bitnami/kubeless/pkg/utils"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

type functionEventType string

type functionEvent struct {
	typ  functionEventType
	spec spec.FunctionSpec
}

type Function struct {
	logger    *logrus.Entry
	kclient   *client.Client
	status    *Status
	Spec      *spec.FunctionSpec
	Name      string
	Namespace string
	eventCh   chan *functionEvent
}

func New(c *client.Client, name, ns string, spec *spec.FunctionSpec, wg *sync.WaitGroup) error {
	return new(c, name, ns, spec, wg)
}

func Delete(c *client.Client, name, ns string, wg *sync.WaitGroup) error {
	return delete(c, name, ns, wg)
}

func new(kclient *client.Client, name, ns string, spec *spec.FunctionSpec, wg *sync.WaitGroup) error {
	f := &Function{
		logger:    logrus.WithField("pkg", "function").WithField("function-name", name),
		kclient:   kclient,
		Name:      name,
		Namespace: ns,
		eventCh:   make(chan *functionEvent, 100),
		Spec:      spec,
		status:    &Status{},
	}

	err := utils.CreateK8sResources(f.Namespace, f.Name, f.Spec, kclient)
	if err != nil {
		return err
	}

	wg.Add(1)

	return nil
}

func delete(kclient *client.Client, name, ns string, wg *sync.WaitGroup) error {
	err := utils.DeleteK8sResources(ns, name, kclient)
	wg.Add(1)
	return err
}
