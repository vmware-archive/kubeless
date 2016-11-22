package controller

import (
	"github.com/skippbox/kubeless/pkg/spec"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type FunctionList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	unversioned.ListMeta `json:"metadata,omitempty"`
	// Items is a list of third party objects
	Items []spec.Function `json:"items"`
}
