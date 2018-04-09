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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTTPTrigger is Kubeless resource representing HTTP trigger event source
type HTTPTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              HTTPTriggerSpec `json:"spec"`
}

// HTTPTriggerSpec defines specification for HTTP trigger
type HTTPTriggerSpec struct {
	FunctionName    string `json:"function-name"` // Name of the associated function
	HostName        string `json:"host-name"`
	TLSAcme         bool   `json:"tls"`
	TLSSecret       string `json:"tls-secret"`
	Path            string `json:"path"`
	BasicAuthSecret string `json:"basic-auth-secret"`
	Gateway         string `json:"gateway"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTTPTriggerList is list of HTTPTrigger's
type HTTPTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is a list of third party objects
	Items []*HTTPTrigger `json:"items"`
}
