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

package event

import (
	"k8s.io/kubernetes/pkg/api"
)

// Event represent an event got from k8s api server
// Events from different endpoints need to be casted to KubelessEvent
// before being able to be handled by handler
type Event struct {
	Namespace string
	Kind      string
	Component string
	Host      string
	Reason    string
	Status    string
	Name      string
}

var m = map[string]string{
	"created": "Normal",
	"deleted": "Danger",
	"updated": "Warning",
}

// New create new KubelessEvent
func New(obj interface{}, action string) Event {
	var namespace, kind, component, host, reason, status, name string
	if apiService, ok := obj.(*api.Service); ok {
		namespace = apiService.ObjectMeta.Namespace
		kind = "service"
		component = string(apiService.Spec.Type)
		reason = action
		status = m[action]
		name = apiService.Name
	} else if apiPod, ok := obj.(*api.Pod); ok {
		namespace = apiPod.ObjectMeta.Namespace
		kind = "pod"
		reason = action
		host = apiPod.Spec.NodeName
		status = m[action]
		name = apiPod.Name
	} else if apiRC, ok := obj.(*api.ReplicationController); ok {
		name = apiRC.TypeMeta.Kind
		kind = "replication controller"
		reason = action
		status = m[action]
		name = apiRC.Name
	}

	kbEvent := Event{
		Namespace: namespace,
		Kind:      kind,
		Component: component,
		Host:      host,
		Reason:    reason,
		Status:    status,
		Name:      name,
	}

	return kbEvent
}
