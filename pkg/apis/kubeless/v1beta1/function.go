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
	"k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Function object
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              FunctionSpec `json:"spec"`
}

// FunctionSpec contains func specification
type FunctionSpec struct {
	Handler                 string                          `json:"handler"`               // Function handler: "file.function"
	Function                string                          `json:"function"`              // Function file content or URL of the function
	FunctionContentType     string                          `json:"function-content-type"` // Function file content type (plain text, base64 or zip)
	Checksum                string                          `json:"checksum"`              // Checksum of the file
	Runtime                 string                          `json:"runtime"`               // Function runtime to use
	Timeout                 string                          `json:"timeout"`               // Maximum timeout for the function to complete its execution
	Deps                    string                          `json:"deps"`                  // Function dependencies
	Deployment              v1beta1.Deployment              `json:"deployment" protobuf:"bytes,3,opt,name=template"`
	ServiceSpec             v1.ServiceSpec                  `json:"service"`
	HorizontalPodAutoscaler v2beta1.HorizontalPodAutoscaler `json:"horizontalPodAutoscaler" protobuf:"bytes,3,opt,name=horizontalPodAutoscaler"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FunctionList contains map of functions
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is a list of third party objects
	Items []*Function `json:"items"`
}
