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

// KinesisTrigger is Kubeless resource representing Kinesis stream as event source
type KinesisTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              KinesisTriggerSpec `json:"spec"`
}

// KinesisTriggerSpec defines specification for KinesisTrigger
type KinesisTriggerSpec struct {
	FunctionName string `json:"function-name"` // Name of the associated function
	Region       string `json:"aws-region"`    // Name of the AWS region corresponding to the stream
	Secret       string `json:"secret"`        // Name of the Kubernetes secret that holds the AWS access key and secret key
	Stream       string `json:"stream"`        // Kinesis Stream name
	ShardID      string `json:"shard"`         // Kinesis Stream shard-id
	Endpoint     string `json:"endpoint"`      // Endpoint url of the Kinesis service
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KinesisTriggerList is list of KinesisTrigger's
type KinesisTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is a list of third party objects
	Items []*KinesisTrigger `json:"items"`
}
