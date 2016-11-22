package spec

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type Function struct {
	unversioned.TypeMeta `json:",inline"`
	api.ObjectMeta       `json:"metadata,omitempty"`
	Spec                 FunctionSpec `json:"spec"`
}

type FunctionSpec struct {
	Handler string `json:"handler"`
	Lambda  string `json:"lambda"`
	Runtime string `json:"runtime"`
	Version string `json:"version"`
}
