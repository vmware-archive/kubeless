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
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/unversioned"
)

// Function object
type Function struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             api.ObjectMeta `json:"metadata"`
	Spec                 FunctionSpec   `json:"spec"`
}

// FunctionSpec contains func specification
type FunctionSpec struct {
	Handler  string `json:"handler"`
	Function string `json:"function"`
	Runtime  string `json:"runtime"`
	Type     string `json:"type"`
	Topic    string `json:"topic"`
	Deps     string `json:"deps"`
}

// FunctionList contains map of functions
type FunctionList struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             unversioned.ListMeta `json:"metadata"`

	// Items is a list of third party objects
	Items []*Function `json:"items"`
}

// GetObjectKind required to satisfy Object interface
func (f *Function) GetObjectKind() unversioned.ObjectKind {
	return &f.TypeMeta
}

// GetObjectMeta required to satisfy ObjectMetaAccessor interface
func (f *Function) GetObjectMeta() meta.Object {
	return &f.Metadata
}

// GetObjectKind required to satisfy Object interface
func (fl *FunctionList) GetObjectKind() unversioned.ObjectKind {
	return &fl.TypeMeta
}

// GetListMeta required to satisfy ListMetaAccessor interface
func (fl *FunctionList) GetListMeta() unversioned.List {
	return &fl.Metadata
}
