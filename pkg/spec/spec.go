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

package spec

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Function object
type Function struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ObjectMeta `json:"metadata"`
	Spec            FunctionSpec      `json:"spec"`
}

// FunctionSpec contains func specification
type FunctionSpec struct {
	Handler  string            `json:"handler"`
	Function string            `json:"function"`
	Runtime  string            `json:"runtime"`
	Type     string            `json:"type"`
	Topic    string            `json:"topic"`
	Deps     string            `json:"deps"`
	Memory   resource.Quantity `json:"memory"`
	Env      map[string]string `json:"env"`
	Desc     string            `json:"description"`
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
