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

package spec

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/autoscaling/v2alpha1"
)

// Function object
type Function struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ObjectMeta `json:"metadata"`
	Spec            FunctionSpec      `json:"spec"`
}

// FunctionSpec contains func specification
type FunctionSpec struct {
	Handler                 string                           `json:"handler"`               // Function handler: "file.function"
	Function                string                           `json:"function"`              // Function file content or URL of the function
	FunctionContentType     string                           `json:"function-content-type"` // Function file content type (plain text, base64 or zip)
	Checksum                string                           `json:"checksum"`              // Checksum of the file
	Runtime                 string                           `json:"runtime"`               // Function runtime to use
	Type                    string                           `json:"type"`                  // Function trigger type
	Topic                   string                           `json:"topic"`                 // Function topic trigger (for PubSub type)
	Schedule                string                           `json:"schedule"`              // Function scheduled time (for Schedule type)
	Timeout                 string                           `json:"timeout"`               // Maximum timeout for the function to complete its execution
	Deps                    string                           `json:"deps"`                  // Function dependencies
	ServiceSpec             v1.ServiceSpec                   `json:"service"`
	Template                v1.PodTemplateSpec               `json:"template" protobuf:"bytes,3,opt,name=template"`
	HorizontalPodAutoscaler v2alpha1.HorizontalPodAutoscaler `json:"horizontalPodAutoscaler" protobuf:"bytes,3,opt,name=horizontalPodAutoscaler"`
}

// FunctionList contains map of functions
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ListMeta `json:"metadata"`

	// Items is a list of third party objects
	Items []*Function `json:"items"`
}

// GetObjectKind required to satisfy Object interface
func (e *Function) GetObjectKind() schema.ObjectKind {
	return &e.TypeMeta
}

// GetObjectMeta required to satisfy ObjectMetaAccessor interface
func (e *Function) GetObjectMeta() metav1.Object {
	return &e.Metadata
}

// GetObjectKind required to satisfy Object interface
func (el *FunctionList) GetObjectKind() schema.ObjectKind {
	return &el.TypeMeta
}

// GetListMeta required to satisfy ListMetaAccessor interface
func (el *FunctionList) GetListMeta() metav1.List {
	return &el.Metadata
}
